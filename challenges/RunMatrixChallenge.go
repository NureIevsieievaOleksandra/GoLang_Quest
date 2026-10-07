package challenges

import (
	"Quest/models"
	"fmt"
)

type Point struct {
	X int
	Y int
}

type Node struct {
	X         int
	Y         int
	DirX      int
	DirY      int
	StepCount int
	Path      []Point
}

func RunMatrixChallenge(state *models.GameState) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 2: МАТРИЦЯ ПАСТОК ===")

	path := findSafePath(10, 10)
	if path == nil {
		fmt.Println("Безпечного шляху не знайдено! Пастка заблокована.")
		return
	}

	fmt.Println("\nБезпечний маршрут знайдено:")
	for i, p := range path {
		if i == len(path)-1 {
			fmt.Printf("(%d,%d)\n", p.X, p.Y)
		} else {
			fmt.Printf("(%d,%d) -> ", p.X, p.Y)
		}
	}

	if !state.Task5Done {
		pointsReward := 7
		airReward := 720.0

		state.Points += pointsReward
		state.AirHours += airReward
		state.Task5Done = true

		fmt.Printf("\nНараховано: +%d балів!\n", pointsReward)
		fmt.Printf("Отримано запас повітря на місяць (+%.0f год)!\n", airReward)
	} else {
		fmt.Println("\n(Маршрут знайдено повторно, бали вже були нараховані раніше)")
	}
}

func isSafe(x, y int) bool {
	if x < 0 || x > 10 || y < 0 || y > 10 {
		return false
	}
	if x == y {
		return true
	}
	return x%3 == 0
}

func findSafePath(targetX, targetY int) []Point {
	directions := [][2]int{
		{1, 0},  // вправо
		{0, 1},  // вниз
		{1, 1},  // діагональ
		{-1, 0}, // вліво
		{0, -1}, // вгору
	}

	visited := make(map[string]bool)

	startNode := Node{
		X:         0,
		Y:         0,
		DirX:      0,
		DirY:      0,
		StepCount: 0,
		Path:      []Point{{0, 0}},
	}

	queue := []Node{startNode}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.X == targetX && curr.Y == targetY {
			return curr.Path
		}

		for _, d := range directions {
			dx, dy := d[0], d[1]
			nextX, nextY := curr.X+dx, curr.Y+dy

			if !isSafe(nextX, nextY) {
				continue
			}

			nextStepCount := 1
			if curr.DirX == dx && curr.DirY == dy {
				nextStepCount = curr.StepCount + 1
			}

			if nextStepCount > 3 {
				continue
			}

			stateKey := fmt.Sprintf("%d,%d,%d,%d,%d", nextX, nextY, dx, dy, nextStepCount)
			if visited[stateKey] {
				continue
			}
			visited[stateKey] = true

			newPath := make([]Point, len(curr.Path)+1)
			copy(newPath, curr.Path)
			newPath[len(curr.Path)] = Point{X: nextX, Y: nextY}

			queue = append(queue, Node{
				X:         nextX,
				Y:         nextY,
				DirX:      dx,
				DirY:      dy,
				StepCount: nextStepCount,
				Path:      newPath,
			})
		}
	}

	return nil
}
