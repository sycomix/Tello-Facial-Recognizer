package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
)

type pair struct {
	x float64
	y float64
}

var leftX, leftY, rightX, rightY atomic.Value

const offset = 32767.0

func main() {
	joystickAdaptor := joystick.NewAdaptor()
	stick := joystick.NewDriver(joystickAdaptor, "xbox360")

	drone := tello.NewDriver("8888")

	work := func() {
		leftX.Store(float64(0.0))
		leftY.Store(float64(0.0))
		rightX.Store(float64(0.0))
		rightY.Store(float64(0.0))

		stick.On(joystick.StartPress, func(data interface{}) {
			fmt.Println("Take off!")
			drone.TakeOff()
		})

		stick.On(joystick.BackPress, func(data interface{}) {
			fmt.Println("Attempting To Land")
			drone.Land()
		})

		stick.On(joystick.UpPress, func(data interface{}) {
			fmt.Println("FrontFlip")
			drone.FrontFlip()
		})

		stick.On(joystick.DownPress, func(data interface{}) {
			fmt.Println("BackFlip")
			drone.BackFlip()
		})

		stick.On(joystick.RightPress, func(data interface{}) {
			fmt.Println("RightFlip")
			drone.RightFlip()
		})

		stick.On(joystick.LeftPress, func(data interface{}) {
			fmt.Println("LeftFlip")
			drone.LeftFlip()
		})

		stick.On(joystick.APress, func(data interface{}) {
			fmt.Println("Ready For Throw Takeoff")
			drone.ThrowTakeOff()
		})

		stick.On(joystick.BPress, func(data interface{}) {
			fmt.Println("Boing!!")
			drone.Bounce()
		})

		stick.On(joystick.LeftX, func(data interface{}) {
			val := float64(data.(int16))
			leftX.Store(val)
		})

		stick.On(joystick.LeftY, func(data interface{}) {
			val := float64(data.(int16))
			leftY.Store(val)
		})

		stick.On(joystick.RightX, func(data interface{}) {
			val := float64(data.(int16))
			rightX.Store(val)
		})

		stick.On(joystick.RightY, func(data interface{}) {
			val := float64(data.(int16))
			rightY.Store(val)
		})

		stick.On(joystick.LBPress, func(data interface{}) {
			fmt.Println("Slow Mode Enabled")
			drone.SetFastMode()
		})

		stick.On(joystick.RBPress, func(data interface{}) {
			fmt.Println("Slow Mode Disabled")
			drone.SetSlowMode()
		})

		gobot.Every(10*time.Millisecond, func() {
			rightStick := getRightStick()

			switch {
			case rightStick.y < -10:
				drone.Forward(tello.ValidatePitch(rightStick.y, offset))
			case rightStick.y > 10:
				drone.Backward(tello.ValidatePitch(rightStick.y, offset))
			default:
				drone.Forward(0)
			}

			switch {
			case rightStick.x > 10:
				drone.Right(tello.ValidatePitch(rightStick.x, offset))
			case rightStick.x < -10:
				drone.Left(tello.ValidatePitch(rightStick.x, offset))
			default:
				drone.Right(0)
			}
		})

		gobot.Every(10*time.Millisecond, func() {
			leftStick := getLeftStick()
			switch {
			case leftStick.y < -10:
				drone.Up(tello.ValidatePitch(leftStick.y, offset))
			case leftStick.y > 10:
				drone.Down(tello.ValidatePitch(leftStick.y, offset))
			default:
				drone.Up(0)
			}

			switch {
			case leftStick.x > 20:
				drone.Clockwise(tello.ValidatePitch(leftStick.x, offset))
			case leftStick.x < -20:
				drone.CounterClockwise(tello.ValidatePitch(leftStick.x, offset))
			default:
				drone.Clockwise(0)
			}
		})
	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{stick, drone},
		work,
	)

	robot.Start()
}

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}
