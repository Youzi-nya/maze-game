package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/danicat/simpleansi"
)

// 变量结构体的定义
var maze []string
var player sprite
var ghosts []*sprite

type sprite struct {
	row int
	col int
}

var score int
var numDots int
var lives = 1

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

	for row, line := range maze {
		for col, char := range line {
			switch char {
			case 'P':
				player = sprite{row, col}
			case 'G':
				ghosts = append(ghosts, &sprite{row, col})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

// 输入读取
func readInput() (string, error) {
	buffer := make([]byte, 100)

	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err
	}

	if cnt == 1 && buffer[0] == 0x1b {
		return "ESC", nil
	} else if cnt >= 3 {
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}
		}
	}

	return "", nil
}

// 处理输入的移动按钮
func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int) {
	newRow, newCol = oldRow, oldCol

	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(maze) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(maze) {
			newRow = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(maze[0]) - 1
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(maze[0]) {
			newCol = 0
		}
	}
	if maze[newRow][newCol] == '#' {
		newCol = oldCol
		newRow = oldRow
	}

	return
}

// 角色移动的实现
func movePlayer(dir string) {
	player.row, player.col = makeMove(player.row, player.col, dir)
	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]
	}
}

// 幽灵AI的实现
func drawDirection() string {
	dir := rand.Intn(4)
	move := map[int]string{
		0: "UP",
		1: "DOWN",
		2: "RIGHT",
		3: "LEFT",
	}
	return move[dir]
}

// 幽灵移动的实现
func moveghosts() {
	for _, g := range ghosts {
		dir := drawDirection()
		g.row, g.col = makeMove(g.row, g.col, dir)
	}
}

// 打印迷宫
func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '#':
				fallthrough
			case '.':
				fmt.Printf("%c", chr)
			default:
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	//打印角色
	simpleansi.MoveCursor(player.row, player.col)
	fmt.Print("P")
	//打印幽灵
	for _, g := range ghosts {
		simpleansi.MoveCursor(g.row, g.col)
		fmt.Print("G")
	}
	//鼠标光标移出
	simpleansi.MoveCursor(len(maze)+1, 0)
	fmt.Println("Score:", score, "\tLives", lives)
}

// 启动Cbreak模式
func initialise() {
	cbTerm := exec.Command("stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("unble to activate cbreak mode!!!", err)
	}
}

// 恢复cooked模式
func cleanup() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("unable to restore cooked mode!!!!", err)
	}
}

// 主函数
func main() {
	//启动cbreak模式
	initialise()
	defer cleanup()

	//读取文件
	err := loadMaze("maze01.txt")
	if err != nil {
		log.Println("failed to load maze:", err)
		return
	}
	input := make(chan string)
	go func(ch chan<- string) {
		for {
			input, err := readInput()
			if err != nil {
				log.Println("error reading input", err)
				ch <- "ESC"
			}
			ch <- input
		}
	}(input)

	//游戏循环
	for {

		//输入
		// input, err := readInput()
		// if err != nil {
		// 	log.Println("error reading input:", err)
		// 	break
		// }

		//移动角色
		select {
		case inp := <-input:
			if inp == "ESC" {
				lives = 0
			}
			movePlayer(inp)
		default:
		}
		//幽灵移动
		moveghosts()

		//游戏结束
		for _, g := range ghosts {
			if player == *g {
				lives--
			}
		}
		//更新屏幕
		printScreen()

		//退出游戏
		if numDots == 0 || lives <= 0 {
			break
		}

		//游戏延时
		time.Sleep(100 * time.Millisecond)
	}
}
