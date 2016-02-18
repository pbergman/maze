package cli

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pbergman/maze/config"
	"github.com/pbergman/maze/builder"
	"github.com/pbergman/maze/solver"
)

func App() {
	var wg sync.WaitGroup
	log.Printf("Building maze with ratio %d, width %d, height: %d", config.Config.Scale, config.Config.Width, config.Config.Height)
	maze := builder.NewMazeImageBuilder(config.Config.Height, config.Config.Width)
	maze.SetRatio(config.Config.Scale)
	log.Printf("Getting matrix and image: %s", maze)
	matrix, err := maze.GetMatrix()
	checkError(err)
	fmt.Println(matrix)
	log.Print("Solving maze")
	start := time.Now()
	walker := solver.NewWalker(matrix)
	walker.Solve()
	log.Printf("Done %s", time.Now().Sub(start))
	fmt.Println(walker)
	wg.Add(3)
	go func() {
		log.Printf("Saving maze: %s", config.Config.Files.Raw)
		f, err := os.Create(config.Config.Files.Raw)
		checkError(err)
		defer f.Close()
		matrix.ToFile(f)
		wg.Done()
	}()
	go func() {
		log.Printf("Saving solved maze: %s", config.Config.Files.Solved)
		o, err := os.Create(config.Config.Files.Solved)
		checkError(err)
		defer o.Close()
		walker.ToFile(o)
		wg.Done()
	}()
	go func() {
		log.Printf("Saving solved animation maze: %s", config.Config.Files.Animation)
		o, err := os.Create(config.Config.Files.Animation)
		checkError(err)
		defer o.Close()
		walker.CreateAnimationImage(o)
		wg.Done()
	}()
	wg.Wait()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}