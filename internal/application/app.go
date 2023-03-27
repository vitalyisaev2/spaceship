package application

import (
	"os"
	"time"

	"github.com/AndreyAD1/spaceship/internal/services"
	"github.com/charmbracelet/log"
)

type Application struct {
	Logger       *log.Logger
	FrameTimeout time.Duration
}

func GetApplication(logger *log.Logger) Application {
	return Application{logger, 10 * time.Millisecond}
}

func (this Application) quit(screenSvc *services.ScreenService) {
	screenSvc.Finish()
	os.Exit(0)
}

func (this Application) Run() error {
	screenService, err := services.GetScreenService()
	if err != nil {
		return err
	}
	defer this.quit(screenService)

	objectChannel := make(chan services.ScreenObject)
	objectsToLoose := []services.ScreenObject{}
	services.GenerateMeteorites(objectChannel)
	services.GenerateShip(screenService, objectChannel)
	go screenService.PollScreenEvents()

	this.Logger.Debug("start an event loop")
	for {
		if screenService.Exit() {
			break
		}
	ObjectLoop:
		for {
			this.Logger.Debugf("get object info. Objects to loose: %v", objectsToLoose)
			select {
			case object := <-objectChannel:
				screenService.Draw(object)
				objectsToLoose = append(objectsToLoose, object)
			default:
				this.Logger.Debugf("channel is empty, object: %v", objectsToLoose)
				for _, object := range objectsToLoose {
					object.Unblock()
				}
				objectsToLoose = objectsToLoose[:0]
				break ObjectLoop
			}
		}
		screenService.ShowScreen()
		time.Sleep(this.FrameTimeout)
		screenService.ClearScreen()
	}
	return nil
}
