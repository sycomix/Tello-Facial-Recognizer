package main

import (
	"fmt"
	"io"
	"os/exec"
	"sync/atomic"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
	"gobot.io/x/gobot/platforms/opencv"
	"gocv.io/x/gocv"
)

type pair struct {
	x float64
	y float64
}

const (
	frameSize = 960 * 720 * 3
)

var leftX, leftY, rightX, rightY atomic.Value

const offset = 32767.0

func main() {
	//_, currentfile, _, _ := runtime.Caller(0)
	//cascade := path.Join(path.Dir(currentfile), "haarcascade_frontalface_alt.xml")

	joystickAdaptor := joystick.NewAdaptor()
	stick := joystick.NewDriver(joystickAdaptor, "xbox360")
	window := opencv.NewWindowDriver()

	drone := tello.NewDriver("8889")

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
			drone.SetSlowMode()
		})

		stick.On(joystick.RBPress, func(data interface{}) {
			fmt.Println("Slow Mode Disabled")
			drone.SetFastMode()
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

		ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-pix_fmt", "bgr24", "-vcodec", "rawvideo",
			"-an", "-sn", "-s", "960x720", "-f", "rawvideo", "pipe:1")
		ffmpegIn, _ := ffmpeg.StdinPipe()
		ffmpegOut, _ := ffmpeg.StdoutPipe()
		if err := ffmpeg.Start(); err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			for {
				buf := make([]byte, frameSize)
				if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
					fmt.Println(err)
					continue
				}

				img := gocv.NewMatFromBytes(720, 960, gocv.MatTypeCV8UC3, buf)
				if img.Empty() {
					continue
				}
				window.ShowImage(img)
				window.WaitKey(1)
			}
		}()

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})
		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				fmt.Println(err)
			}
		})
	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{stick, drone, window},
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
