package service

import (
	"EIP/model"
	"bufio"
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func NewCalibrate(unixAddr, klipperPath string, messageChan chan string, cTime int64) (bool, string) {
	// new input shaper calibrate
	log.Println("start calibrate")
	conn, err := net.DialTimeout("unix", unixAddr, 2*time.Second)
	if err != nil {

		messageChan <- "0%  " + err.Error()
		return false, "Can't Connect to Socket!"
	}
	// Reader
	go func() {
		reader := bufio.NewReader(conn)
		// totally 292 lines
		line := 0
		xFilename := ""
		yFilename := ""
		scriptPath := klipperPath + "scripts/calibrate_shaper.py"
		for {
			line = line + 1
			currentPercent := strconv.Itoa(line * 100 / 300)
			message, err := reader.ReadString('\x03')
			if err != nil {
				log.Println("error reading:", err)
				conn.Close()
				return
			}
			clearMessage := strings.TrimRight(message, "\x03")

			var response map[string]interface{}
			err = json.Unmarshal([]byte(clearMessage), &response)
			if err != nil {
				messageChan <- "0%  " + err.Error()
				conn.Close()
				return
			}

			if errorData, ok := response["error"]; ok {
				errorMessage := errorData.(map[string]interface{})["message"].(string)
				messageChan <- "0%  " + errorMessage
				conn.Close()
				return
			} else if paramsData, ok := response["params"]; ok {
				responseMessage := paramsData.(map[string]interface{})["response"].(string)
				messageChan <- currentPercent + "%  " + responseMessage

				// find csv and save_config str
				if xFilename == "" {
					xFilename = getFileName(responseMessage)
				} else {
					if yFilename == "" {
						yFilename = getFileName(responseMessage)

					} else {
						messageChan <- "100% Shaper Calibrate Complete, Generate Img Now..."
						// Generate x image
						outPath := "./static/result/" + int64ToStr(cTime) + "x.png"
						cmd := exec.Command(scriptPath, xFilename, "-o", outPath, "-c", "-v")
						stdoutPipe, err := cmd.StdoutPipe()
						err = cmd.Start()
						if err != nil {
							messageChan <- "0%  " + err.Error()
							log.Println(err)
							return
						}
						reader := bufio.NewReader(stdoutPipe)
						for {
							line, err := reader.ReadString('\n')
							if err != nil {
								break
							}
							messageChan <- "100%...  " + line
						}
						err = cmd.Wait()
						if err != nil {
							messageChan <- "0%  " + err.Error()
							log.Println(err)
							return
						}
						// Generate y image
						outPath = "./static/result/" + int64ToStr(cTime) + "y.png"
						cmd = exec.Command(scriptPath, yFilename, "-o", outPath, "-c", "-v")
						stdoutPipe, err = cmd.StdoutPipe()
						err = cmd.Start()
						if err != nil {
							messageChan <- err.Error()
							return
						}
						reader = bufio.NewReader(stdoutPipe)
						for {
							line, err := reader.ReadString('\n')
							if err != nil {
								break
							}
							messageChan <- "100%...  " + line
						}
						err = cmd.Wait()
						if err != nil {
							messageChan <- "0%  " + err.Error()
							return
						}
						messageChan <- "100% Command completed"
						nRecord := model.Record{
							Model: gorm.Model{},
							XAxis: xFilename,
							YAxis: yFilename,
							Time:  cTime,
							Name:  time.Unix(cTime, 0).Format("2006-01-02 15:04"),
						}
						model.NewRecord(nRecord)
						conn.Close()
						return
					}

				}

			} else {
				messageChan <- "0%  " + message
			}

		}
	}()

	// reg endpoint
	message := `{"id": 1, "method": "gcode/subscribe_output", "params": {"response_template":{}}}`
	data := []byte(message)
	// 向服务器发送消息
	_, err = conn.Write(append(data, 0x03))
	if err != nil {
		return false, "Send Message Fail!"
	}
	// Run SHAPER_CALIBRATE
	message = `{"id": 2, "method": "gcode/script", "params": {"script": "SHAPER_CALIBRATE"}}`
	data = []byte(message)
	// 向服务器发送消息
	_, err = conn.Write(append(data, 0x03))
	if err != nil {
		return false, "Send Message Fail!"
	}
	return true, ""
}

func getFileName(str string) string {
	filePrefix := "/tmp/"
	fileName := ""

	if strings.Contains(str, filePrefix) {
		responseParts := strings.Split(str, " ")
		for _, part := range responseParts {
			if strings.HasPrefix(part, filePrefix) {
				fileName = part
				break
			}
		}
	}

	if fileName != "" {
		return fileName
	} else {
		return ""
	}

}

func int64ToStr(input int64) string {
	str := strconv.FormatInt(input, 10)
	return str
}
