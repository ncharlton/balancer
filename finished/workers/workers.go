package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"image/color"
	"log"
	"math/rand"
	"time"
)

const (
	screenWidth, screenHeight = 1000, 200
	rectangleSize = 10
	workerCount = 1
	requestCount = 100
)

var (
	workers [workerCount]*Worker
	green = color.RGBA{10, 255, 50, 255}
	blue = color.RGBA{85, 51, 237, 1}
	yellow = color.RGBA{245, 229,27, 1}
	red = color.RGBA{240, 52, 52, 1}
	orange = color.RGBA{248, 148, 6, 1}
	grey = color.RGBA{240, 240, 214, 1}

	colors = [5]color.RGBA{
		green,
		blue,
		yellow,
		red,
		orange,
	}
)

type Position struct {
	x, y int
}

type Request struct {
	fn func() int
	color color.RGBA
}

type Worker struct {
	id int
	position Position
	request Request
	color color.RGBA
}

func (w *Worker) work(requestChannel chan Request) {
	for {
		req := <- requestChannel
		w.request = req
		w.color = w.request.color
		w.request.fn()
		w.color = grey
	}
}

type Display struct {}

func (d *Display) Update() error {
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	for _, worker := range workers {
		for i := 0; i < rectangleSize; i++ {
			screen.Set(worker.position.x + i, worker.position.y, worker.color)
			screen.Set(worker.position.x, worker.position.y + i, worker.color)
			for j := 0; j < rectangleSize; j++ {
				screen.Set(worker.position.x, worker.position.y + i, worker.color)
				screen.Set(worker.position.x + i, worker.position.y + j, worker.color)
			}
		}
	}
}

func (d *Display) Layout(outsideWith, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWith, outsideHeight
}

func main() {
	requestChannel := make(chan Request)

	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			id: i,
			color: grey,
			position: Position{
				x: i * 50 + 50,
				y: 100,
			},
		}

		workers[i] = worker

		fmt.Printf("worker %d initialized \n", worker.id)

		go worker.work(requestChannel)
	}

	go triggerRequests(requestChannel)

	display := &Display{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Workers")

	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}

func triggerRequests(requestChannel chan Request) {
	time.Sleep(time.Second * 2)
	start := time.Now()

	for i := 0; i < requestCount; i++ {
		req := Request{
			color: colors[rand.Intn(5 - 0)],
			fn: doSomeWork,
		}

		fmt.Printf("request %d init \n", i)

		requestChannel <- req
	}

	elapsed := time.Since(start)
	log.Printf("Requests took %s", elapsed)
}

func doSomeWork() int {
	n := rand.Int63n(1e9)
	time.Sleep(time.Duration(n))
	return int(n)
}