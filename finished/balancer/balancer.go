package main

import (
	"container/heap"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"image/color"
	"log"
	"math/rand"
	"time"
)

const requestCount = 60
const workerCount = 10
const screenWidth, screenHeight = 600, 200

var (
	balancer *Balancer

	green1 = color.RGBA{85, 153, 56, 1}
	green2 = color.RGBA{93, 186, 54, 1}
	green3 = color.RGBA{177, 201, 54, 1}
	orange1 = color.RGBA{201, 184, 54, 1}
	orange2 = color.RGBA{226, 176, 60, 1}
	orange3 = color.RGBA{230, 149, 55, 1}
	red1 = color.RGBA{230, 99, 55, 1}
	red2 = color.RGBA{230, 79, 55, 1}
	red3 = color.RGBA{227, 13, 13, 1}

	colors = [9]color.RGBA{
		green1,
		green2,
		green3,
		orange1,
		orange2,
		orange3,
		red1,
		red2,
		red3,
	}
)

type Position struct {
	x int
	y int
}

type Request struct {
	fn func() int
	c  chan int
}

type Worker struct {
	i        int
	requests chan Request
	pending  int
	position Position
}

type Display struct{}

func (d *Display) Update() error {
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	for _, worker := range balancer.pool {
		col := red3

		if worker.pending < 9 {
			col = colors[worker.pending]
		}

		for i := 0; i < 10; i++ {
			screen.Set(worker.position.x + i, worker.position.y, col)
			screen.Set(worker.position.x, worker.position.y + i, col)
			for j := 0; j < 10; j++ {
				screen.Set(worker.position.x, worker.position.y + i, col)
				screen.Set(worker.position.x + i, worker.position.y + j, col)
			}
		}
	}
}

func (d *Display) Layout(outsideWith, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWith, outsideHeight
}

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].i = i
	a[j].i = j
}

func (p *Pool) Push(x interface{}) {
	a := *p
	n := len(a)
	a = a[0 : n+1]
	w := x.(*Worker)
	a[n] = w
	w.i = n
	*p = a
}

func (p *Pool) Pop() interface{} {
	a := *p
	*p = a[0 : len(a)-1]
	w := a[len(a)-1]
	w.i = -1
	return w
}

type Balancer struct {
	pool Pool
	done chan *Worker
}

func NewBalancer() *Balancer {
	done := make(chan *Worker, workerCount)
	balancer := &Balancer{
		pool: make(Pool, 0, workerCount),
		done: done,
	}
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			requests: make(chan Request, requestCount),
			position: Position{
				x: i * 50 + 50,
				y: 100,
			},
		}

		heap.Push(&balancer.pool, worker)
		go worker.work(balancer.done)
	}

	return balancer
}

func (w *Worker) work(done chan *Worker) {
	for {
		req := <-w.requests
		req.c <- req.fn()
		done <- w
	}
}

// trigger requests
func triggerRequests(work chan Request) {
	c := make(chan int)
	for {
		time.Sleep(time.Duration(rand.Int63n(workerCount * 2e9)))
		work <- Request{doSomeWork, c}
		<-c
	}
}

func (b *Balancer) balance(work chan Request) {
	for {
		select {
		case request := <-work:
			b.dispatch(request)
		case worker := <-b.done:
			b.completed(worker)
		}
		b.print()
	}
}

// dispatch request to worker
func (b *Balancer) dispatch(request Request) {
	worker := heap.Pop(&b.pool).(*Worker)
	worker.requests <- request
	worker.pending++
	heap.Push(&b.pool, worker)
}

// worker completed request.
func (b *Balancer) completed(worker *Worker) {
	worker.pending--
	heap.Remove(&b.pool, worker.i)
	heap.Push(&b.pool, worker)
}

// request work functionn
func doSomeWork() int {
	n := rand.Int63n(1e9)
	time.Sleep(time.Duration(workerCount * n))
	return int(n)
}

func main() {
	work := make(chan Request)
	for i := 0; i < requestCount; i++ {
		go triggerRequests(work)
	}

	b := NewBalancer()

	go b.balance(work)

	// display
	balancer = b
	display := &Display{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Load balancer")

	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}

// print worker statistics
func (b *Balancer) print() {
	sum := 0
	sumsq := 0
	for _, w := range b.pool {
		fmt.Printf("%d ", w.pending)
		sum += w.pending
		sumsq += w.pending * w.pending
	}

	avg := float64(sum) / float64(len(b.pool))
	fmt.Printf(" %.2f\n", avg)
}
