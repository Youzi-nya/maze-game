package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

var maze []string

// 读取迷宫文件
func loadMaze(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	return nil
}

// 打印迷宫
func printScreen() {
	for _, line := range maze {
		fmt.Println(line)
	}
}

func main() {
	err := loadMaze("maze01.txt")
	if err != nil {
		log.Println("failed to load maze:", err)
		return
	}

	for {
		printScreen()
		break
	}

}
