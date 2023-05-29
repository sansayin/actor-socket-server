package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
)

type (
	MSGLogin struct{ 
        user string
        pass string
    }

	ServiceActor struct {
		listener       net.Listener
		sessions map[SessionNew]struct{}
	}

	SessionNew struct {
		conn *net.Conn
	}

	SessionExit struct {
		session SessionNew
	}
)

func newSession(conn *net.Conn) actor.Producer {
	return func() actor.Actor {
		return &SessionNew{
			conn: conn,
		}
	}
}

func (this *SessionNew) Receive(context actor.Context) {

	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Println("Sessions Start", msg)
		context.Send(context.Parent(), this)
		go this.sessionHandler(context)
	case *actor.Stopping:
		(*this.conn).Close()
	case *actor.Stopped:
		log.Println("Connection Closed")
	}
}

func (this *SessionNew) sessionHandler(context actor.Context) {
	buf := make([]byte, 1024)
	for {
		n, err := (*this.conn).Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("\nconn read error:%v", err)
			break
		}
		reply := strings.ToUpper(string(buf[0:n]))
		(*this.conn).Write([]byte(reply))
	}
	log.Printf("\nSession %v Handler exit\n", this)
	context.Send(context.Parent(), &SessionExit{session: *this})
}

func newService() actor.Actor {
	nl, err := net.Listen("tcp", ":8899")
	if err != nil {
		log.Fatal(err)
	}
	return &ServiceActor{
		listener:       nl,
		sessions: make(map[SessionNew]struct{}),
	}
}

func (this *ServiceActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Println("Listening :8899")
		go this.hanlder(context)
	case *SessionNew:
		log.Printf("Add Session: %v\n", msg.conn)
		this.sessions[*msg] = struct{}{}
	case *SessionExit:
		delete(this.sessions, msg.session)
		log.Printf("Sessions NO.: %v\n", len(this.sessions))
	case *MSGLogin:
		log.Printf("%v Says:Hello %v\n", context.Self().GetAddress(), msg.user)
	}
}

func (this *ServiceActor) hanlder(context actor.Context) {
	for {
		conn, _ := this.listener.Accept()
		log.Printf("New Connection: %v\n", conn)
		props := actor.PropsFromProducer(newSession(&conn))
		pid, err := context.SpawnNamed(props, fmt.Sprintf("session/%v", conn))
		if err != nil {
			log.Println(pid, err)
		}
	}
}

func init() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Running Routins: %v\n", runtime.NumGoroutine())
			}
		}
	}()
}
func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(newService)
	system.Root.Spawn(props)

	_, _ = console.ReadLine()
}
