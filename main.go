package main

import (
	"fmt"
	"pipeline/pipeline"
)

func main() {
	input := make(chan int)
	done := make(chan struct{})
	go pipeline.Read(input, done)

	p := pipeline.NewPipeLineInt(done, pipeline.NegativeFilterStage, pipeline.NotDividedFilterStage, pipeline.BufferStage)
	output := p.Run(input)

	for {
		select {
		case data, ok := <-output:
			if !ok {
				return
			}
			fmt.Println("Обработанные данные:", data)
		case <-done:
			return
		}
	}
}
