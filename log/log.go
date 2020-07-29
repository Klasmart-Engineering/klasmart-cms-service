package log

import (
	"fmt"
	"time"
)

type ILog interface{
	Debug(text string)
	Debugf(text string, args... interface{})
	Info(text string)
	Infof(text string, args... interface{})
	Warn(text string)
	Warnf(text string, args... interface{})
	Error(text string)
	Errorf(text string, args... interface{})
}

type StdLog struct {

}

var log *StdLog

func init(){
	log = new(StdLog)
}

func (s *StdLog) format(text string, level string) string{
	return time.Now().Format("2006-01-02 15:04:05") + " [" + level + "] " + text + "\n"
}

func (s *StdLog) Debug(text string){
	fmt.Print(s.format(text, "DEBUG"))
}
func (s *StdLog) Debugf(text string, args... interface{}){
	fmt.Printf(s.format(text, "DEBUG"), args...)
}
func (s *StdLog) Info(text string){
	fmt.Print(s.format(text, "INFO"))
}
func (s *StdLog) Infof(text string, args... interface{}){
	fmt.Printf(s.format(text, "INFO"), args...)
}
func (s *StdLog) Warn(text string){
	fmt.Print(s.format(text, "WARN"))
}
func (s *StdLog) Warnf(text string, args... interface{}){
	fmt.Printf(s.format(text, "WARN"), args...)
}
func (s *StdLog) Error(text string){
	fmt.Print(s.format(text, "ERROR"))
}
func (s *StdLog) Errorf(text string, args... interface{}){
	fmt.Printf(s.format(text, "ERROR"), args...)
}

func Get() *StdLog{
	return log
}