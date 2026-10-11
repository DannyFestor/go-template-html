package spawn

type Spawner struct{}

func Start() {
	go work()
}

func (s *Spawner) Run() {
	go work()
}

func work() {}
