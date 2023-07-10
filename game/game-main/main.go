package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/danicat/simpleansi"
)

// 变量结构体的定义
var (
	configFile = flag.String("config-file", "config.json", "path to custom configuration file")
	mazeFile   = flag.String("maze-file", "maze01.txt", "path to a custom maze file")
)
var maze []string
var player sprite
var ghosts []*sprite

type sprite struct {
	row int
	col int
}
type config struct {
	Player   string `json:"player"`
	Ghost    string `json:"ghost"`
	Wall     string `json:"wall"`
	Dot      string `json:"dot"`
	Pill     string `json:"pill"`
	Death    string `json:"death"`
	Space    string `json:"space"`
	UseEmoji bool   `json:"use_emoji"`
}

var score int
var numDots int
var lives = 1

var cfg config

// 读入config解码json
func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

// 调整水平位移
func moveCursor(row, col int) {
	if cfg.UseEmoji {
		simpleansi.MoveCursor(row, col*2)
	} else {
		simpleansi.MoveCursor(row, col)
	}
}

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
				player = sprite{row: row, col: col}
			case 'G':
				ghosts = append(ghosts, &sprite{row: row, col: col})
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

	// removeDot := func(row, col int){

	// }
	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]
	case 'X':
		score += 10
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
				fmt.Print(simpleansi.WithBlueBackground(cfg.Wall))
			case '.':
				fmt.Print(cfg.Dot)
			case 'X':
				fmt.Print(cfg.Pill)
			default:
				fmt.Print(cfg.Space)
			}
		}
		fmt.Println()
	}
	//打印角色
	moveCursor(player.row, player.col)
	fmt.Print(cfg.Player)
	//打印幽灵
	for _, g := range ghosts {
		moveCursor(g.row, g.col)
		fmt.Print(cfg.Ghost)
	}
	//鼠标光标移出
	moveCursor(len(maze)+1, 0)
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
	flag.Parse()
	//启动cbreak模式
	initialise()
	defer cleanup()

	//读取文件
	err := loadMaze(*mazeFile)
	if err != nil {
		log.Println("failed to load maze:", err)
		return
	}

	//解析json
	err = loadConfig(*configFile)
	if err != nil {
		log.Println("failed to load configuration:", err)
	}

	//输入
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
			if lives == 0 {
				moveCursor(player.row, player.col)
				fmt.Print(cfg.Death)
				moveCursor(len(maze)+2, 0)
			}
			break
		}

		//游戏延时
		time.Sleep(100 * time.Millisecond)
	}
}
