package challenges

import (
	"Quest/survival"
	"fmt"
)

type Point struct {
	x int
	y int
}

type Node struct {
	x         int
	y         int
	dirX      int
	dirY      int
	stepCount int
	path      []Point
}

func RunMatrixChallenge(state *survival.GameState) {
	fmt.Println("\n=== ВИПРОБУВАННЯ 2: МАТРИЦЯ ПАСТОК ===")

	path := findSafePath(10, 10)
	if path == nil {
		fmt.Println("Безпечного шляху не знайдено! Пастка заблокована.")
		return
	}

	fmt.Println("\nБезпечний маршрут знайдено:")
	for i, p := range path {
		if i == len(path)-1 {
			fmt.Printf("(%d,%d)\n", p.x, p.y)
		} else {
			fmt.Printf("(%d,%d) -> ", p.x, p.y)
		}
	}

	if !state.Task5Done {
		pointsReward := 7
		airReward := 720.0

		survival.ApplyEvent(state,
			survival.AddPoints(pointsReward),
			survival.AddSupplies(survival.DaySupplies{Air: airReward}),
			func(s *survival.Survivor) { s.Task5Done = true },
		)

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

func findSafePath(targetx, targety int) []Point {
	directions := [][2]int{
		{1, 0},  // вправо
		{0, 1},  // вниз
		{1, 1},  // діагональ
		{-1, 0}, // вліво
		{0, -1}, // вгору
	}

	visited := make(map[string]bool)

	startNode := Node{
		x:         0,
		y:         0,
		dirX:      0,
		dirY:      0,
		stepCount: 0,
		path:      []Point{{0, 0}},
	}

	queue := []Node{startNode}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.x == targetx && curr.y == targety {
			return curr.path
		}

		for _, d := range directions {
			dx, dy := d[0], d[1]
			nextX, nextY := curr.x+dx, curr.y+dy

			if !isSafe(nextX, nextY) {
				continue
			}

			nextstepCount := 1
			if curr.dirX == dx && curr.dirY == dy {
				nextstepCount = curr.stepCount + 1
			}

			if nextstepCount > 3 {
				continue
			}

			stateKey := fmt.Sprintf("%d,%d,%d,%d,%d", nextX, nextY, dx, dy, nextstepCount)
			if visited[stateKey] {
				continue
			}
			visited[stateKey] = true

			newpath := make([]Point, len(curr.path)+1)
			copy(newpath, curr.path)
			newpath[len(curr.path)] = Point{x: nextX, y: nextY}

			queue = append(queue, Node{
				x:         nextX,
				y:         nextY,
				dirX:      dx,
				dirY:      dy,
				stepCount: nextstepCount,
				path:      newpath,
			})
		}
	}

	return nil
}
