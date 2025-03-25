package main

import (
	"math/rand"

	ecs "github.com/BrownNPC/simple-ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Define Components
type Position rl.Vector2
type Velocity rl.Vector2

type Tag int

const (
	Player Tag = iota
	Rock
)

var lastSpawnTime float32 = 0

// for convenience sake
var player ecs.Entity
var gameOver bool

func main() {
	rl.InitWindow(360, 640, "Dodge the Rocks")
	defer rl.CloseWindow()

	p := ecs.New(100)

	// Create the player
	player = ecs.NewEntity(p)
	ecs.Add3(p, player,
		Position{X: 180, Y: 600}, // Center bottom
		Velocity{X: 0, Y: 0},
		Player,
	)

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		Update(p, dt)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		renderingSystem(p)
		rl.EndDrawing()
	}
}

func Update(p *ecs.Pool, dt float32) {
	if !gameOver {
		playerMovementSystem(p)
		movementSystem(p, dt)
		collisionSystem(p, player)
		spawnSystem(p, dt)
		despawnSystem(p)
	}
}

func playerMovementSystem(p *ecs.Pool) {
	POSITION, VELOCITY, TAG :=
		ecs.GetStorage3[Position, Velocity, Tag](p)

	for _, e := range POSITION.And(VELOCITY, TAG) {
		pos, vel := POSITION.Get(e), VELOCITY.Get(e)
		tag := TAG.Get(e)
		if tag != Player {
			continue
		}

		// Move left
		if rl.IsKeyDown(rl.KeyLeft) {
			vel.X = -200
		} else if rl.IsKeyDown(rl.KeyRight) {
			vel.X = 200
		} else {
			vel.X = 0
		}

		POSITION.Update(e, pos)
		VELOCITY.Update(e, vel)
	}
}

func movementSystem(p *ecs.Pool, dt float32) {
	POSITION, VELOCITY, TAG :=
		ecs.GetStorage3[Position, Velocity, Tag](p)

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
	POSITION, TAG :=
		ecs.GetStorage2[Position, Tag](p)
	for _, e := range POSITION.And(TAG) {
		if TAG.Get(e) == Player {
			continue
		}
		rock_pos := POSITION.Get(e)
		plr_pos := POSITION.Get(player)

		if rl.CheckCollisionCircleRec(rl.Vector2(rock_pos), 12,
			rl.NewRectangle(plr_pos.X, plr_pos.Y, 30, 30)) {
			gameOver = true
		}
	}
}
func renderingSystem(p *ecs.Pool) {
	if gameOver {
		rl.DrawText("GAME OVER", (360/2)-rl.MeasureText("GAME OVER", 30)/2, 640/2, 30, rl.Red)
		return
	}

	POSITION, TAG :=
		ecs.GetStorage2[Position, Tag](p)

	// Draw rocks and player
	for _, e := range POSITION.And(TAG) {
		pos, tag :=
			POSITION.Get(e), TAG.Get(e)
		if tag == Rock {
			pos := POSITION.Get(e)
			rl.DrawCircle(int32(pos.X), int32(pos.Y), 12, rl.DarkBrown)
		} else if tag == Player {
			rl.DrawRectangle(int32(pos.X-15), int32(pos.Y-15), 30, 30, rl.Blue)
		}
	}
}
