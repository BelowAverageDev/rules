package rules

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func sortPoints(p []Point) {
	sort.Slice(p, func(i, j int) bool {
		if p[i].X != p[j].X {
			return p[i].X < p[j].X
		}
		return p[i].Y < p[j].Y
	})
}

func TestCreateDefaultBoardState(t *testing.T) {
	tests := []struct {
		Height          int32
		Width           int32
		IDs             []string
		ExpectedNumFood int
		Err             error
	}{
		{1, 1, []string{"one"}, 0, nil},
		{1, 2, []string{"one"}, 0, nil},
		{1, 4, []string{"one"}, 1, nil},
		{2, 2, []string{"one"}, 1, nil},
		{9, 8, []string{"one"}, 1, nil},
		{2, 2, []string{"one", "two"}, 0, nil},
		{1, 1, []string{"one", "two"}, 2, ErrorNoRoomForSnake},
		{1, 2, []string{"one", "two"}, 2, ErrorNoRoomForSnake},
		{BoardSizeSmall, BoardSizeSmall, []string{"one", "two"}, 3, nil},
	}

	for testNum, test := range tests {
		state, err := CreateDefaultBoardState(test.Width, test.Height, test.IDs)
		require.Equal(t, test.Err, err)
		if err != nil {
			require.Nil(t, state)
			continue
		}
		require.NotNil(t, state)
		require.Equal(t, test.Width, state.Width)
		require.Equal(t, test.Height, state.Height)
		require.Equal(t, len(test.IDs), len(state.Snakes))
		for i, id := range test.IDs {
			require.Equal(t, id, state.Snakes[i].ID)
		}
		require.Len(t, state.Food, test.ExpectedNumFood, testNum)
		require.Len(t, state.Hazards, 0, testNum)
	}
}

func TestPlaceSnakesDefault(t *testing.T) {
	// Because placement is random, we only test to ensure
	// that snake bodies are populated correctly
	// Note: because snakes are randomly spawned on even diagonal points, the board can accomodate number of snakes equal to: width*height/2
	tests := []struct {
		BoardState *BoardState
		SnakeIDs   []string
		Err        error
	}{
		{
			&BoardState{
				Width:  1,
				Height: 1,
			},
			make([]string, 1),
			nil,
		},
		{
			&BoardState{
				Width:  1,
				Height: 1,
			},
			make([]string, 2),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  2,
				Height: 1,
			},
			make([]string, 2),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  1,
				Height: 2,
			},
			make([]string, 2),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  10,
				Height: 5,
			},
			make([]string, 24),
			nil,
		},
		{
			&BoardState{
				Width:  5,
				Height: 10,
			},
			make([]string, 25),
			nil,
		},
		{
			&BoardState{
				Width:  10,
				Height: 5,
			},
			make([]string, 49),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  5,
				Height: 10,
			},
			make([]string, 50),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  25,
				Height: 2,
			},
			make([]string, 51),
			ErrorNoRoomForSnake,
		},
		{
			&BoardState{
				Width:  BoardSizeSmall,
				Height: BoardSizeSmall,
			},
			make([]string, 1),
			nil,
		},
		{
			&BoardState{
				Width:  BoardSizeSmall,
				Height: BoardSizeSmall,
			},
			make([]string, 8),
			nil,
		},
		{
			&BoardState{
				Width:  BoardSizeSmall,
				Height: BoardSizeSmall,
			},
			make([]string, 9),
			ErrorTooManySnakes,
		},
		{
			&BoardState{
				Width:  BoardSizeMedium,
				Height: BoardSizeMedium,
			},
			make([]string, 8),
			nil,
		},
		{
			&BoardState{
				Width:  BoardSizeMedium,
				Height: BoardSizeMedium,
			},
			make([]string, 9),
			ErrorTooManySnakes,
		},
		{
			&BoardState{
				Width:  BoardSizeLarge,
				Height: BoardSizeLarge,
			},
			make([]string, 8),
			nil,
		},
		{
			&BoardState{
				Width:  BoardSizeLarge,
				Height: BoardSizeLarge,
			},
			make([]string, 9),
			ErrorTooManySnakes,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.BoardState.Width, test.BoardState.Height, len(test.SnakeIDs)), func(t *testing.T) {
			require.Equal(t, test.BoardState.Width*test.BoardState.Height, int32(len(getUnoccupiedPoints(test.BoardState, true))))
			err := PlaceSnakesAutomatically(test.BoardState, test.SnakeIDs)
			require.Equal(t, test.Err, err, "Snakes: %d", len(test.BoardState.Snakes))
			if err == nil {
				for i := 0; i < len(test.BoardState.Snakes); i++ {
					require.Len(t, test.BoardState.Snakes[i].Body, 3)
					for _, point := range test.BoardState.Snakes[i].Body {
						require.GreaterOrEqual(t, point.X, int32(0))
						require.GreaterOrEqual(t, point.Y, int32(0))
						require.Less(t, point.X, test.BoardState.Width)
						require.Less(t, point.Y, test.BoardState.Height)
					}

					for j := 0; j < len(test.BoardState.Snakes); j++ {
						if j == i {
							continue
						}
						require.NotEqual(t, test.BoardState.Snakes[j].Body[0], test.BoardState.Snakes[i].Body[0], "Snakes placed at same square")
					}

					// All snakes are expected to be placed on an even square - this is true even of fixed positions for known board sizes
					var snakePlacedOnEvenSquare bool = ((test.BoardState.Snakes[i].Body[0].X + test.BoardState.Snakes[i].Body[0].Y) % 2) == 0
					require.Equal(t, true, snakePlacedOnEvenSquare)
				}
			}
		})
	}
}

func TestPlaceSnake(t *testing.T) {
	// TODO: Should PlaceSnake check for boundaries?
	boardState := NewBoardState(BoardSizeSmall, BoardSizeSmall)
	require.Empty(t, boardState.Snakes)

	_ = PlaceSnake(boardState, "a", []Point{{0, 0}, {1, 0}, {1, 1}})

	require.Len(t, boardState.Snakes, 1)
	require.Equal(t, Snake{
		ID:              "a",
		Body:            []Point{{0, 0}, {1, 0}, {1, 1}},
		Health:          SnakeMaxHealth,
		EliminatedCause: NotEliminated,
		EliminatedBy:    "",
	}, boardState.Snakes[0])

	_ = PlaceSnake(boardState, "b", []Point{{0, 2}, {1, 2}, {3, 2}})

	require.Len(t, boardState.Snakes, 2)
	require.Equal(t, Snake{
		ID:              "b",
		Body:            []Point{{0, 2}, {1, 2}, {3, 2}},
		Health:          SnakeMaxHealth,
		EliminatedCause: NotEliminated,
		EliminatedBy:    "",
	}, boardState.Snakes[1])
}

func TestPlaceFood(t *testing.T) {
	tests := []struct {
		BoardState   *BoardState
		ExpectedFood int
	}{
		{
			&BoardState{
				Width:  1,
				Height: 1,
				Snakes: make([]Snake, 1),
			},
			1,
		},
		{
			&BoardState{
				Width:  1,
				Height: 2,
				Snakes: make([]Snake, 2),
			},
			2,
		},
		{
			&BoardState{
				Width:  101,
				Height: 202,
				Snakes: make([]Snake, 17),
			},
			17,
		},
		{
			&BoardState{
				Width:  10,
				Height: 20,
				Snakes: make([]Snake, 305),
			},
			200,
		},
		{
			&BoardState{
				Width:  BoardSizeSmall,
				Height: BoardSizeSmall,
				Snakes: []Snake{
					{Body: []Point{{5, 1}}},
					{Body: []Point{{5, 3}}},
					{Body: []Point{{5, 5}}},
				},
			},
			4, // +1 because of fixed spawn locations
		},
		{
			&BoardState{
				Width:  BoardSizeMedium,
				Height: BoardSizeMedium,
				Snakes: []Snake{
					{Body: []Point{{1, 1}}},
					{Body: []Point{{1, 5}}},
					{Body: []Point{{1, 9}}},
					{Body: []Point{{5, 1}}},
					{Body: []Point{{5, 9}}},
					{Body: []Point{{9, 1}}},
					{Body: []Point{{9, 5}}},
					{Body: []Point{{9, 9}}},
				},
			},
			9, // +1 because of fixed spawn locations
		},
		{
			&BoardState{
				Width:  BoardSizeLarge,
				Height: BoardSizeLarge,
				Snakes: []Snake{
					{Body: []Point{{1, 1}}},
					{Body: []Point{{1, 9}}},
					{Body: []Point{{1, 17}}},
					{Body: []Point{{17, 1}}},
					{Body: []Point{{17, 9}}},
					{Body: []Point{{17, 17}}},
				},
			},
			7, // +1 because of fixed spawn locations
		},
	}

	for _, test := range tests {
		require.Len(t, test.BoardState.Food, 0)
		err := PlaceFoodAutomatically(test.BoardState)
		require.NoError(t, err)
		require.Equal(t, test.ExpectedFood, len(test.BoardState.Food))
		for _, point := range test.BoardState.Food {
			require.GreaterOrEqual(t, point.X, int32(0))
			require.GreaterOrEqual(t, point.Y, int32(0))
			require.Less(t, point.X, test.BoardState.Width)
			require.Less(t, point.Y, test.BoardState.Height)
		}
	}
}

func TestPlaceFoodFixed(t *testing.T) {
	tests := []struct {
		BoardState *BoardState
	}{
		{
			&BoardState{
				Width:  BoardSizeSmall,
				Height: BoardSizeSmall,
				Snakes: []Snake{
					{Body: []Point{{1, 3}}},
				},
			},
		},
		{
			&BoardState{
				Width:  BoardSizeMedium,
				Height: BoardSizeMedium,
				Snakes: []Snake{
					{Body: []Point{{1, 1}}},
					{Body: []Point{{1, 5}}},
					{Body: []Point{{9, 5}}},
					{Body: []Point{{9, 9}}},
				},
			},
		},
		{
			&BoardState{
				Width:  BoardSizeLarge,
				Height: BoardSizeLarge,
				Snakes: []Snake{
					{Body: []Point{{1, 1}}},
					{Body: []Point{{1, 9}}},
					{Body: []Point{{1, 17}}},
					{Body: []Point{{9, 1}}},
					{Body: []Point{{9, 17}}},
					{Body: []Point{{17, 1}}},
					{Body: []Point{{17, 9}}},
					{Body: []Point{{17, 17}}},
				},
			},
		},
	}

	for _, test := range tests {
		require.Len(t, test.BoardState.Food, 0)

		err := PlaceFoodFixed(test.BoardState)
		require.NoError(t, err)
		require.Equal(t, len(test.BoardState.Snakes)+1, len(test.BoardState.Food))

		midPoint := Point{(test.BoardState.Width - 1) / 2, (test.BoardState.Height - 1) / 2}

		// Make sure every snake has food within 2 moves of it
		for _, snake := range test.BoardState.Snakes {
			head := snake.Body[0]

			bottomLeft := Point{head.X - 1, head.Y - 1}
			topLeft := Point{head.X - 1, head.Y + 1}
			bottomRight := Point{head.X + 1, head.Y - 1}
			topRight := Point{head.X + 1, head.Y + 1}

			foundFoodInTwoMoves := false
			for _, food := range test.BoardState.Food {
				if food == bottomLeft || food == topLeft || food == bottomRight || food == topRight {
					foundFoodInTwoMoves = true
					// Ensure it's not closer to the center than snake
					require.True(t, getDistanceBetweenPoints(head, midPoint) <= getDistanceBetweenPoints(food, midPoint))
					break
				}
			}
			require.True(t, foundFoodInTwoMoves)
		}

		// Make sure one food exists in center of board
		foundFoodInCenter := false
		for _, food := range test.BoardState.Food {
			if food == midPoint {
				foundFoodInCenter = true
				break
			}
		}
		require.True(t, foundFoodInCenter)
	}
}

func TestPlaceFoodFixedNoRoom(t *testing.T) {
	boardState := &BoardState{
		Width:  3,
		Height: 3,
		Snakes: []Snake{
			{Body: []Point{{1, 1}}},
		},
		Food: []Point{},
	}
	err := PlaceFoodFixed(boardState)
	require.Error(t, err)
}

func TestPlaceFoodFixedNoRoom_Corners(t *testing.T) {
	boardState := &BoardState{
		Width:  7,
		Height: 7,
		Snakes: []Snake{
			{Body: []Point{{1, 1}}},
			{Body: []Point{{1, 5}}},
			{Body: []Point{{5, 1}}},
			{Body: []Point{{5, 5}}},
		},
		Food: []Point{},
	}

	// There are only three possible spawn locations for each snake,
	// so repeat calls to place food should fail after 3 successes
	err := PlaceFoodFixed(boardState)
	require.NoError(t, err)
	boardState.Food = boardState.Food[:len(boardState.Food)-1] // Center food
	require.Equal(t, 4, len(boardState.Food))

	err = PlaceFoodFixed(boardState)
	require.NoError(t, err)
	boardState.Food = boardState.Food[:len(boardState.Food)-1] // Center food
	require.Equal(t, 8, len(boardState.Food))

	err = PlaceFoodFixed(boardState)
	require.NoError(t, err)
	boardState.Food = boardState.Food[:len(boardState.Food)-1] // Center food
	require.Equal(t, 12, len(boardState.Food))

	// And now there should be no more room.
	err = PlaceFoodFixed(boardState)
	require.Error(t, err)

	expectedFood := []Point{
		{0, 0}, {0, 2}, {2, 0}, // Snake @ {1, 1}
		{0, 4}, {0, 6}, {2, 6}, // Snake @ {1, 5}
		{4, 0}, {6, 0}, {6, 2}, // Snake @ {5, 1}
		{4, 6}, {6, 4}, {6, 6}, // Snake @ {5, 5}
	}
	sortPoints(expectedFood)
	sortPoints(boardState.Food)
	require.Equal(t, expectedFood, boardState.Food)
}

func TestPlaceFoodFixedNoRoom_Cardinal(t *testing.T) {
	boardState := &BoardState{
		Width:  11,
		Height: 11,
		Snakes: []Snake{
			{Body: []Point{{1, 5}}},
			{Body: []Point{{5, 1}}},
			{Body: []Point{{5, 9}}},
			{Body: []Point{{9, 5}}},
		},
		Food: []Point{},
	}

	// There are only two possible spawn locations for each snake,
	// so repeat calls to place food should fail after 2 successes
	err := PlaceFoodFixed(boardState)
	require.NoError(t, err)
	boardState.Food = boardState.Food[:len(boardState.Food)-1] // Center food
	require.Equal(t, 4, len(boardState.Food))

	err = PlaceFoodFixed(boardState)
	require.NoError(t, err)
	boardState.Food = boardState.Food[:len(boardState.Food)-1] // Center food
	require.Equal(t, 8, len(boardState.Food))

	// And now there should be no more room.
	err = PlaceFoodFixed(boardState)
	require.Error(t, err)

	expectedFood := []Point{
		{0, 4}, {0, 6}, // Snake @ {1, 5}
		{4, 0}, {6, 0}, // Snake @ {5, 1}
		{4, 10}, {6, 10}, // Snake @ {5, 9}
		{10, 4}, {10, 6}, // Snake @ {9, 5}
	}
	sortPoints(expectedFood)
	sortPoints(boardState.Food)
	require.Equal(t, expectedFood, boardState.Food)
}

func TestGetDistanceBetweenPoints(t *testing.T) {
	tests := []struct {
		A        Point
		B        Point
		Expected int32
	}{
		{Point{0, 0}, Point{0, 0}, 0},
		{Point{0, 0}, Point{1, 0}, 1},
		{Point{0, 0}, Point{0, 1}, 1},
		{Point{0, 0}, Point{1, 1}, 2},
		{Point{0, 0}, Point{4, 4}, 8},
		{Point{0, 0}, Point{4, 6}, 10},
		{Point{8, 0}, Point{8, 0}, 0},
		{Point{8, 0}, Point{8, 8}, 8},
		{Point{8, 0}, Point{0, 8}, 16},
	}

	for _, test := range tests {
		require.Equal(t, getDistanceBetweenPoints(test.A, test.B), test.Expected)
		require.Equal(t, getDistanceBetweenPoints(test.B, test.A), test.Expected)
	}
}

func TestIsKnownBoardSize(t *testing.T) {
	tests := []struct {
		Width    int32
		Height   int32
		Expected bool
	}{
		{1, 1, false},
		{0, 0, false},
		{0, 45, false},
		{45, 1, false},
		{7, 7, true},
		{11, 11, true},
		{19, 19, true},
		{7, 11, false},
		{11, 19, false},
		{19, 7, false},
	}

	for _, test := range tests {
		result := isKnownBoardSize(&BoardState{Width: test.Width, Height: test.Height})
		require.Equal(t, test.Expected, result)
	}
}

func TestGetUnoccupiedPoints(t *testing.T) {
	tests := []struct {
		Board    *BoardState
		Expected []Point
	}{
		{
			&BoardState{
				Height: 1,
				Width:  1,
			},
			[]Point{{0, 0}},
		},
		{
			&BoardState{
				Height: 1,
				Width:  2,
			},
			[]Point{{0, 0}, {1, 0}},
		},
		{
			&BoardState{
				Height: 1,
				Width:  1,
				Food:   []Point{{0, 0}, {101, 202}, {-4, -5}},
			},
			[]Point{},
		},
		{
			&BoardState{
				Height: 2,
				Width:  2,
				Food:   []Point{{0, 0}, {1, 0}},
			},
			[]Point{{0, 1}, {1, 1}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  2,
				Food:   []Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			},
			[]Point{},
		},
		{
			&BoardState{
				Height: 4,
				Width:  1,
				Snakes: []Snake{
					{Body: []Point{{0, 0}}},
				},
			},
			[]Point{{0, 1}, {0, 2}, {0, 3}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  3,
				Snakes: []Snake{
					{Body: []Point{{0, 0}, {1, 0}, {1, 1}}},
				},
			},
			[]Point{{0, 1}, {2, 0}, {2, 1}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  3,
				Food:   []Point{{0, 0}, {1, 0}, {1, 1}, {2, 0}},
				Snakes: []Snake{
					{Body: []Point{{0, 0}, {1, 0}, {1, 1}}},
					{Body: []Point{{0, 1}}},
				},
			},
			[]Point{{2, 1}},
		},
	}

	for _, test := range tests {
		unoccupiedPoints := getUnoccupiedPoints(test.Board, true)
		require.Equal(t, len(test.Expected), len(unoccupiedPoints))
		for i, e := range test.Expected {
			require.Equal(t, e, unoccupiedPoints[i])
		}
	}
}

func TestGetEvenUnoccupiedPoints(t *testing.T) {
	tests := []struct {
		Board    *BoardState
		Expected []Point
	}{
		{
			&BoardState{
				Height: 1,
				Width:  1,
			},
			[]Point{{0, 0}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  2,
			},
			[]Point{{0, 0}, {1, 1}},
		},
		{
			&BoardState{
				Height: 1,
				Width:  1,
				Food:   []Point{{0, 0}, {101, 202}, {-4, -5}},
			},
			[]Point{},
		},
		{
			&BoardState{
				Height: 2,
				Width:  2,
				Food:   []Point{{0, 0}, {1, 0}},
			},
			[]Point{{1, 1}},
		},
		{
			&BoardState{
				Height: 4,
				Width:  4,
				Food:   []Point{{0, 0}, {0, 2}, {1, 1}, {1, 3}, {2, 0}, {2, 2}, {3, 1}, {3, 3}},
			},
			[]Point{},
		},
		{
			&BoardState{
				Height: 4,
				Width:  1,
				Snakes: []Snake{
					{Body: []Point{{0, 0}}},
				},
			},
			[]Point{{0, 2}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  3,
				Snakes: []Snake{
					{Body: []Point{{0, 0}, {1, 0}, {1, 1}}},
				},
			},
			[]Point{{2, 0}},
		},
		{
			&BoardState{
				Height: 2,
				Width:  3,
				Food:   []Point{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
				Snakes: []Snake{
					{Body: []Point{{0, 0}, {1, 0}, {1, 1}}},
					{Body: []Point{{0, 1}}},
				},
			},
			[]Point{{2, 0}},
		},
	}

	for _, test := range tests {
		evenUnoccupiedPoints := getEvenUnoccupiedPoints(test.Board)
		require.Equal(t, len(test.Expected), len(evenUnoccupiedPoints))
		for i, e := range test.Expected {
			require.Equal(t, e, evenUnoccupiedPoints[i])
		}
	}
}

func TestPlaceFoodRandomly(t *testing.T) {
	b := &BoardState{
		Height: 1,
		Width:  3,
		Snakes: []Snake{
			{Body: []Point{{1, 0}}},
		},
	}
	// Food should never spawn, no room
	err := PlaceFoodRandomly(b, 99)
	require.NoError(t, err)
	require.Equal(t, len(b.Food), 0)
}
