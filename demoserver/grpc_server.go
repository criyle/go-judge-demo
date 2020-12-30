package main

import (
	"bytes"
	"context"
	"time"

	execpb "github.com/criyle/go-judge/pb"
	"github.com/criyle/go-judger-demo/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type demoServer struct {
	pb.UnimplementedDemoBackendServer
	db     *db
	logger *zap.Logger
	client execpb.ExecutorClient

	submit chan *pb.JudgeClientRequest
	update chan *pb.JudgeClientResponse

	register   chan *observer
	unregister chan *observer
	observers  map[*observer]bool
}

func newDemoServer(db *db, client execpb.ExecutorClient, logger *zap.Logger) *demoServer {
	ds := &demoServer{
		db:         db,
		logger:     logger,
		client:     client,
		submit:     make(chan *pb.JudgeClientRequest, 64),
		update:     make(chan *pb.JudgeClientResponse, 64),
		register:   make(chan *observer, 64),
		unregister: make(chan *observer, 64),
		observers:  make(map[*observer]bool),
	}
	go ds.updateLoop()
	return ds
}

type observer struct {
	update chan *pb.JudgeUpdate
}

func (s *demoServer) Submission(ctx context.Context, req *pb.SubmissionRequest) (*pb.SubmissionResponse, error) {
	id := req.GetId()
	m, err := s.db.Query(ctx, id)
	if err != nil {
		return nil, err
	}
	sub := make([]*pb.Submission, 0, len(m))
	for _, v := range m {
		sub = append(sub, &pb.Submission{
			Id:        v.ID.Hex(),
			Language:  convertLanguage(v.Lang),
			Source:    v.Source,
			Date:      timestamppb.New(*v.Date),
			Status:    v.Status,
			TotalTime: v.TotalTime,
			MaxMemory: v.MaxMemory,
			Results:   convertResults(v.Results),
		})
	}
	return &pb.SubmissionResponse{Submissions: sub}, nil
}

func (s *demoServer) Submit(ctx context.Context, req *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	m, err := s.db.Add(ctx, &ClientSubmit{
		Lang:   convertLanguagePB(req.GetLanguage()),
		Source: req.GetSource(),
	})
	if err != nil {
		return nil, err
	}
	s.submit <- &pb.JudgeClientRequest{
		Id:       m.ID.Hex(),
		Language: req.GetLanguage(),
		Source:   req.GetSource(),
	}
	s.update <- &pb.JudgeClientResponse{
		Id:       m.ID.Hex(),
		Language: req.Language,
		Date:     timestamppb.New(*m.Date),
		Source:   m.Source,
	}
	s.logger.Sugar().Debug("submit: ", m)
	return &pb.SubmitResponse{
		Id: m.ID.Hex(),
	}, nil
}

func (s *demoServer) Judge(js pb.DemoBackend_JudgeServer) error {
	for {
		// Send request to client
		req, ok := <-s.submit
		s.logger.Sugar().Info("judge request: ", req)
		if !ok {
			break
		}
		err := js.Send(req)
		if err != nil {
			// If encouters error, do not consume this
			s.submit <- req
			return err
		}
		// Recv updates from client
		for {
			resp, err := js.Recv()
			s.logger.Sugar().Info("judge response: ", req, err)
			if err != nil {
				// If encouters error, do not consume this
				s.submit <- req
				return err
			}
			s.update <- resp
			if resp.Type == "finished" {
				break
			}
		}
	}
	return nil
}

func (s *demoServer) Updates(_ *emptypb.Empty, us pb.DemoBackend_UpdatesServer) error {
	ob := &observer{update: make(chan *pb.JudgeUpdate, 64)}
	s.register <- ob
	defer func() { s.unregister <- ob }()
	for {
		select {
		case <-us.Context().Done():
			return nil

		case u, ok := <-ob.update:
			if !ok {
				return nil
			}
			s.logger.Sugar().Info("send updates:", u)
			if err := us.Send(u); err != nil {
				return err
			}
		}
	}
}

func (s *demoServer) Shell(ss pb.DemoBackend_ShellServer) error {
	ctx, cancel := context.WithCancel(ss.Context())
	defer cancel()

	sc, err := s.client.ExecStream(ctx)
	if err != nil {
		return err
	}
	err = sc.Send(&execpb.StreamRequest{
		Request: &execpb.StreamRequest_ExecRequest{
			ExecRequest: &execpb.Request{
				Cmd: []*execpb.Request_CmdType{{
					Args: []string{"/bin/bash"},
					Env:  []string{"PATH=/usr/local/bin:/usr/bin:/bin", "HOME=/w", "TERM=xterm-256color"},
					Files: []*execpb.Request_File{
						{File: &execpb.Request_File_StreamIn{StreamIn: &execpb.Request_StreamInput{Name: "i"}}},
						{File: &execpb.Request_File_StreamOut{StreamOut: &execpb.Request_StreamOutput{Name: "o"}}},
						{File: &execpb.Request_File_StreamOut{StreamOut: &execpb.Request_StreamOutput{Name: "o"}}},
					},
					Tty:          true,
					CPULimit:     uint64(30 * time.Second),
					RealCPULimit: uint64(30 * time.Minute),
					MemoryLimit:  256 << 20,
					ProcLimit:    50,
				}},
			},
		},
	})
	if err != nil {
		return err
	}
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	go func() {
		defer cancel()

		for {
			msg, err := sc.Recv()
			s.logger.Sugar().Debug("sc recv: ", msg)
			if err != nil {
				return
			}
			switch msg := msg.Response.(type) {
			case *execpb.StreamResponse_ExecOutput:
				output.Write(msg.ExecOutput.Content)
				err = ss.Send(&pb.ShellOutput{Content: msg.ExecOutput.Content})
				if err != nil {
					return
				}

			case *execpb.StreamResponse_ExecResponse:
				err = ss.Send(&pb.ShellOutput{Content: []byte(msg.ExecResponse.String())})
				if err != nil {
					return
				}
				return
			}
		}
	}()

	go func() {
		defer cancel()

		for {
			msg, err := ss.Recv()
			s.logger.Sugar().Debug("ss recv: ", msg)
			if err != nil {
				return
			}
			switch msg := msg.Request.(type) {
			case *pb.ShellInput_Input:
				input.Write(msg.Input.Content)
				err = sc.Send(&execpb.StreamRequest{Request: &execpb.StreamRequest_ExecInput{ExecInput: &execpb.StreamRequest_Input{
					Name: "i", Content: msg.Input.Content,
				}}})
				if err != nil {
					return
				}

			case *pb.ShellInput_Resize:
				err = sc.Send(&execpb.StreamRequest{Request: &execpb.StreamRequest_ExecResize{ExecResize: &execpb.StreamRequest_Resize{
					Name: "i",
					Rows: msg.Resize.Rows,
					Cols: msg.Resize.Cols,
					X:    msg.Resize.X,
					Y:    msg.Resize.Y,
				}}})
			}
		}
	}()
	<-ctx.Done()

	s.db.Store(context.TODO(), &ShellStore{
		Stdin:  input.String(),
		Stdout: output.String(),
	})
	return nil
}

func (s *demoServer) updateLoop() {
	for {
		select {
		case o := <-s.register:
			s.observers[o] = true
		case o := <-s.unregister:
			delete(s.observers, o)

		case u := <-s.update:
			up := &pb.JudgeUpdate{
				Id:       u.GetId(),
				Type:     u.GetType(),
				Status:   u.GetStatus(),
				Date:     u.GetDate(),
				Language: u.GetLanguage(),
				Results:  u.GetResults(),
				Source:   u.GetSource(),
			}
			// save to db
			id, _ := primitive.ObjectIDFromHex(u.GetId())
			s.db.Update(context.TODO(), &JudgerUpdate{
				ID:      &id,
				Type:    u.GetType(),
				Status:  u.GetStatus(),
				Results: convertResultsPB(u.GetResults()),
			})
			// broadcast
			for o := range s.observers {
				select {
				case o.update <- up:
				default:
					// too slow
					//close(o.update)
				}
			}
		}
	}
}

func convertLanguagePB(l *pb.Language) Language {
	return Language{
		Name:           l.GetName(),
		SourceFileName: l.GetSourceFileName(),
		CompileCmd:     l.GetCompileCmd(),
		Executables:    l.GetExecutables(),
		RunCmd:         l.GetRunCmd(),
	}
}

func convertLanguage(l Language) *pb.Language {
	return &pb.Language{
		Name:           l.Name,
		SourceFileName: l.SourceFileName,
		CompileCmd:     l.CompileCmd,
		Executables:    l.Executables,
		RunCmd:         l.RunCmd,
	}
}

func convertResultsPB(r []*pb.Result) []Result {
	rt := make([]Result, 0, len(r))
	for _, v := range r {
		rt = append(rt, convertResultPB(v))
	}
	return rt
}

func convertResultPB(r *pb.Result) Result {
	return Result{
		Time:   r.Time,
		Memory: r.Memory,
		Stdin:  r.Stdin,
		Stdout: r.Stdout,
		Stderr: r.Stderr,
		Log:    r.Log,
	}
}

func convertResults(r []Result) []*pb.Result {
	rt := make([]*pb.Result, 0, len(r))
	for _, v := range r {
		rt = append(rt, convertResult(v))
	}
	return rt
}

func convertResult(r Result) *pb.Result {
	return &pb.Result{
		Time:   r.Time,
		Memory: r.Memory,
		Stdin:  r.Stdin,
		Stdout: r.Stdout,
		Stderr: r.Stderr,
		Log:    r.Log,
	}
}
