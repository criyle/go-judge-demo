package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/criyle/go-judge/judger"
	"github.com/criyle/go-judge/runner"
	"github.com/criyle/go-judge/taskqueue/channel"
	"github.com/criyle/go-sandbox/container"
	"github.com/criyle/go-sandbox/pkg/cgroup"
	"github.com/criyle/go-sandbox/pkg/mount"
)

const (
	envWebURL = "WEB_URL"
)

var (
	compileEnv = []string{
		pathEnv,
		"GOCACHE=/tmp",
	}
	runEnv = []string{pathEnv}
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func init() {
	container.Init()
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	var wg sync.WaitGroup

	c := newClient(os.Getenv(envWebURL), 3*time.Second)

	done := make(chan struct{})
	root, err := ioutil.TempDir("", "dm")
	if err != nil {
		panic(err)
	}
	q := channel.New()
	m, err := mount.NewBuilder().
		// basic exec and lib
		WithBind("/bin", "bin", true).
		WithBind("/lib", "lib", true).
		WithBind("/lib64", "lib64", true).
		WithBind("/usr", "usr", true).
		// java wants /proc/self/exe as it need relative path for lib
		// however, /proc gives interface like /proc/1/fd/3 ..
		// it is fine since open that file will be a EPERM
		// changing the fs uid and gid would be a good idea
		WithProc().
		// some compiler have multiple version
		WithBind("/etc/alternatives", "etc/alternatives", true).
		// fpc wants /etc/fpc.cfg
		WithBind("/etc/fpc.cfg", "etc/fpc.cfg", true).
		// go wants /dev/null
		WithBind("/dev/null", "dev/null", false).
		// work dir
		WithTmpfs("w", "size=8m,nr_inodes=4k").
		// tmp dir
		WithTmpfs("tmp", "size=8m,nr_inodes=4k").
		// finished
		Build(true)

	if err != nil {
		panic(err)
	}
	b := &container.Builder{
		Root:          root,
		Mounts:        m,
		CredGenerator: newCredGen(),
		Stderr:        true,
	}
	cgb, err := cgroup.NewBuilder("go-judger").WithCPUAcct().WithMemory().WithPids().FilterByEnv()
	if err != nil {
		panic(err)
	}
	log.Printf("Initialized cgroup: %v", cgb)
	r := &runner.Runner{
		Builder:       b,
		Queue:         q,
		CgroupBuilder: cgb,
		Language:      &dumbLang{},
	}
	const parallism = 4
	for i := 0; i < parallism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Loop(done)
		}()
	}

	j := &judger.Judger{
		Client:  c,
		Sender:  q,
		Builder: &dumbBuilder{},
	}
	go j.Loop(done)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("interrupted")
	close(done)
	wg.Wait()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

type credGen struct {
	cur uint32
}

func newCredGen() *credGen {
	return &credGen{cur: 10000}
}

func (c *credGen) Get() syscall.Credential {
	n := atomic.AddUint32(&c.cur, 1)
	return syscall.Credential{
		Uid: n,
		Gid: n,
	}
}
