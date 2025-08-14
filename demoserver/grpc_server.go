package main

import (
	"bytes"
	"context"
	"time"

	"github.com/criyle/go-judge-demo/pb"
	execpb "github.com/criyle/go-judge/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		id := v.ID.Hex()
		sub = append(sub, pb.Submission_builder{
			Id:        &id,
			Language:  convertLanguage(v.Lang),
			Source:    &v.Source,
			Date:      timestamppb.New(*v.Date),
			Status:    &v.Status,
			TotalTime: &v.TotalTime,
			MaxMemory: &v.MaxMemory,
			Results:   convertResults(v.Results),
		}.Build())
	}
	return pb.SubmissionResponse_builder{Submissions: sub}.Build(), nil
}

func (s *demoServer) Submit(ctx context.Context, req *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	m, err := s.db.Add(ctx, &ClientSubmit{
		Lang:   convertLanguagePB(req.GetLanguage()),
		Source: req.GetSource(),
	})
	if err != nil {
		return nil, err
	}
	id := m.ID.Hex()
	source := req.GetSource()
	s.submit <- pb.JudgeClientRequest_builder{
		Id:          &id,
		Language:    req.GetLanguage(),
		Source:      &source,
		InputAnswer: req.GetInputAnswer(),
	}.Build()
	s.update <- pb.JudgeClientResponse_builder{
		Id:       &id,
		Language: req.GetLanguage(),
		Date:     timestamppb.New(*m.Date),
		Source:   &source,
	}.Build()
	s.logger.Debug("submit", zap.Any("model", m))
	return pb.SubmitResponse_builder{
		Id: &id,
	}.Build(), nil
}

func (s *demoServer) Judge(js pb.DemoBackend_JudgeServer) error {
	for {
		// Send request to client
		req, ok := <-s.submit
		s.logger.Info("judge request", zap.Any("request", req))
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
			s.logger.Info("judge response", zap.Any("request", req), zap.Error(err))
			if err != nil {
				// If encouters error, do not consume this
				s.submit <- req
				return err
			}
			s.update <- resp
			if resp.GetType() == "finished" {
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
			s.logger.Info("send updates", zap.Any("updates", u))
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
	err = sc.Send(execpb.StreamRequest_builder{
		ExecRequest: execpb.Request_builder{
			Cmd: []*execpb.Request_CmdType{execpb.Request_CmdType_builder{
				Args: []string{"/bin/bash"},
				Env:  []string{"PATH=/usr/local/bin:/usr/bin:/bin", "HOME=/w", "TERM=xterm-256color"},
				Files: []*execpb.Request_File{
					execpb.Request_File_builder{StreamIn: &emptypb.Empty{}}.Build(),
					execpb.Request_File_builder{StreamOut: &emptypb.Empty{}}.Build(),
					execpb.Request_File_builder{StreamOut: &emptypb.Empty{}}.Build(),
				},
				Tty:            true,
				CpuTimeLimit:   uint64(30 * time.Second),
				ClockTimeLimit: uint64(30 * time.Minute),
				MemoryLimit:    256 << 20,
				ProcLimit:      50,
			}.Build()},
		}.Build(),
	}.Build())
	if err != nil {
		return err
	}
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	go func() {
		defer cancel()

		for {
			msg, err := sc.Recv()
			s.logger.Debug("sc recv", zap.Any("message", msg))
			if err != nil {
				return
			}
			switch msg.WhichResponse() {
			case execpb.StreamResponse_ExecOutput_case:
				output.Write(msg.GetExecOutput().GetContent())
				err = ss.Send(pb.ShellOutput_builder{Content: msg.GetExecOutput().GetContent()}.Build())
				if err != nil {
					return
				}

			case execpb.StreamResponse_ExecResponse_case:
				err = ss.Send(pb.ShellOutput_builder{Content: []byte(msg.GetExecResponse().String())}.Build())
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
			s.logger.Debug("ss recv", zap.Any("message", msg))
			if err != nil {
				return
			}
			switch msg.WhichRequest() {
			case pb.ShellInput_Input_case:
				input.Write(msg.GetInput().GetContent())
				err = sc.Send(execpb.StreamRequest_builder{ExecInput: execpb.StreamRequest_Input_builder{
					Content: msg.GetInput().GetContent(),
				}.Build()}.Build())
				if err != nil {
					return
				}

			case pb.ShellInput_Resize_case:
				err = sc.Send(execpb.StreamRequest_builder{ExecResize: execpb.StreamRequest_Resize_builder{
					Rows: msg.GetResize().GetRows(),
					Cols: msg.GetResize().GetCols(),
					X:    msg.GetResize().GetX(),
					Y:    msg.GetResize().GetY(),
				}.Build()}.Build())
				if err != nil {
					return
				}
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
			up := pb.JudgeUpdate_builder{
				Date:     u.GetDate(),
				Language: u.GetLanguage(),
				Results:  u.GetResults(),
			}.Build()
			up.SetId(u.GetId())
			up.SetType(u.GetType())
			up.SetStatus(u.GetStatus())
			up.SetSource(u.GetSource())
			// save to db
			id, _ := bson.ObjectIDFromHex(u.GetId())
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
	return pb.Language_builder{
		Name:           &l.Name,
		SourceFileName: &l.SourceFileName,
		CompileCmd:     &l.CompileCmd,
		Executables:    &l.Executables,
		RunCmd:         &l.RunCmd,
	}.Build()
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
		Time:   r.GetTime(),
		Memory: r.GetMemory(),
		Stdin:  r.GetStdin(),
		Stdout: r.GetStdout(),
		Stderr: r.GetStderr(),
		Log:    r.GetLog(),
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
	return pb.Result_builder{
		Time:   &r.Time,
		Memory: &r.Memory,
		Stdin:  &r.Stdin,
		Stdout: &r.Stdout,
		Stderr: &r.Stderr,
		Log:    &r.Log,
	}.Build()
}
