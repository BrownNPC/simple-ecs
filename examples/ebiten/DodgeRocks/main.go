package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"

	ecs "github.com/BrownNPC/simple-ecs"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// Define Components
type Position struct {
	X, Y float32
}
type Velocity struct {
	X, Y float32
}

type Tag int

const (
	Player Tag = iota
	Rock
)

var lastSpawnTime float32 = 0

// For convenience sake
var player ecs.Entity
var gameOver bool

var playerImage *ebiten.Image
var rockImage *ebiten.Image
var fontFace font.Face = basicfont.Face7x13

func initImages() {
	// Create a blue square for the player (30x30)
	playerImage = ebiten.NewImage(30, 30)
	playerImage.Fill(color.RGBA{0, 0, 255, 255})

	// Create a dark brown circle for the rock (diameter 24, radius 12)
	rockImage = ebiten.NewImage(24, 24)
	r := float32(12)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			dx := float32(x) - r
			dy := float32(y) - r
			if dx*dx+dy*dy <= r*r {
				rockImage.Set(x, y, color.RGBA{101, 67, 33, 255})
			} else {
				rockImage.Set(x, y, color.Transparent)
			}
		}
	}
}

func main() {
	initImages()

	pool := ecs.New(100)

	// Create the player
	player = ecs.NewEntity(pool)
	ecs.Add3(pool, player,
		Position{X: 180, Y: 600}, // Center bottom
		Velocity{X: 0, Y: 0},
		Player,
	)

	game := &Game{pool: pool}
	ebiten.SetWindowSize(360, 640)
	ebiten.SetWindowTitle("Dodge the Rocks")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	pool *ecs.Pool
}

func (g *Game) Update() error {
	var dt float32 = 1.0 / 60.0
	if !gameOver {
		playerMovementSystem(g.pool)
		movementSystem(g.pool, dt)
		collisionSystem(g.pool, player)
		spawnSystem(g.pool, dt)
		despawnSystem(g.pool)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen with white (similar to rl.ClearBackground(rl.RayWhite))
	screen.Fill(color.RGBA{255, 255, 255, 255})
	renderingSystem(g.pool, screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 360, 640
}

func playerMovementSystem(p *ecs.Pool) {
	POSITION, VELOCITY, TAG := ecs.GetStorage3[Position, Velocity, Tag](p)

	for _, e := range POSITION.And(VELOCITY, TAG) {
		pos, vel := POSITION.Get(e), VELOCITY.Get(e)
		tag := TAG.Get(e)
		if tag != Player {
			continue
		}

		// Move left/right using ebiten key input
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			vel.X = -200
		} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			vel.X = 200
		} else {
			vel.X = 0
		}

		POSITION.Update(e, pos)
		VELOCITY.Update(e, vel)
	}
}

func movementSystem(p *ecs.Pool, dt float32) {
	POSITION, VELOCITY, TAG := ecs.GetStorage3[Position, Velocity, Tag](p)

	for _, e := range POSITION.And(VELOCITY) {
		pos, vel := POSITION.Get(e), VELOCITY.Get(e)
		pos.X += vel.X * dt
		pos.Y += vel.Y * dt

		// Clamp player within screen bounds
		if TAG.Get(e) == Player {
			if pos.X < 0 {
				pos.X = 0
			} else if pos.X > 360 {
				pos.X = 360
			}
		}

		POSITION.Update(e, pos)
	}
}

func spawnSystem(p *ecs.Pool, dt float32) {
	lastSpawnTime += dt
	if lastSpawnTime < 0.5 {
		return
	}
	lastSpawnTime = 0 // Reset timer

	e := ecs.NewEntity(p)
	ecs.Add3(p, e,
		Position{X: float32(rand.Intn(360)), Y: 0}, // Random X at top
		Velocity{X: 0, Y: 100},                     // Falling down
		Rock,
	)
}

func despawnSystem(p *ecs.Pool) {
	TAG, POSITION := ecs.GetStorage2[Tag, Position](p)
	for _, e := range TAG.And() {
		if TAG.Get(e) == Rock {
			pos := POSITION.Get(e)
			if pos.Y > 700 {
				ecs.Kill(p, e)
			}
		}
	}
}

func collisionSystem(p *ecs.Pool, player ecs.Entity) {
	POSITION, TAG := ecs.GetStorage2[Position, Tag](p)
	for _, e := range POSITION.And(TAG) {
		if TAG.Get(e) == Player {
			continue
		}
		rockPos := POSITION.Get(e)
		plrPos := POSITION.Get(player)

		if checkCollisionCircleRect(rockPos, 12, plrPos.X, plrPos.Y, 30, 30) {
			gameOver = true
		}
	}
}

func checkCollisionCircleRect(circlePos Position, radius float32, rectX, rectY, rectW, rectH float32) bool {
	// Find the closest point to the circle within the rectangle
	closestX := math.Max(float64(rectX), math.Min(float64(circlePos.X), float64(rectX+rectW)))
	closestY := math.Max(float64(rectY), math.Min(float64(circlePos.Y), float64(rectY+rectH)))
	dx := float64(circlePos.X) - closestX
	dy := float64(circlePos.Y) - closestY

	return dx*dx+dy*dy <= float64(radius*radius)
}

func renderingSystem(p *ecs.Pool, screen *ebiten.Image) {
	if gameOver {
		msg := "GAME OVER"
		bounds := text.BoundString(fontFace, msg)
		textWidth := bounds.Dx()
		x := (360 - textWidth) / 2
		y := 640 / 2
		text.Draw(screen, msg, fontFace, x, y, color.RGBA{255, 0, 0, 255})
		return
	}

	POSITION, TAG := ecs.GetStorage2[Position, Tag](p)

	// Draw rocks and player
	for _, e := range POSITION.And(TAG) {
		pos, tag := POSITION.Get(e), TAG.Get(e)
		if tag == Rock {
			op := &ebiten.DrawImageOptions{}
			// Center the circle image (radius 12)
			op.GeoM.Translate(float64(pos.X-12), float64(pos.Y-12))
			screen.DrawImage(rockImage, op)
		} else if tag == Player {
			op := &ebiten.DrawImageOptions{}
			// Center the player rectangle (half of 30 is 15)
			op.GeoM.Translate(float64(pos.X-15), float64(pos.Y-15))
			screen.DrawImage(playerImage, op)
		}
	}
}
