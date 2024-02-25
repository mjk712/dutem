package dutem

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/amdf/ixxatvci3/candev"
)

/*
Эмуляция ДУТ-ЭМ
всего 8 датчиков

в программе dutecan_emu на Lua было так:

если i - номер датчика от 1 до 8, то:
msg.id = 0x0CF60664 + i
msg.byte1_2 = sensors[i].ctrl.get_level() * 10000 -- m => 0.1 mm
msg.byte7 = sensors[i].ctrl.get_temper() + 40  -- 0 == -40 °C

посылать раз в секунду
если эмулируется несколько датчиков, то все посылать одновременно
*/

const baseDUT = uint32(0x0CF60664)
const numSensors = 8

//FuelParams параметры эмуляции для датчика
type FuelParams struct {
	Enabled     bool
	Level       float64 //(min 0.0 max 1.0)
	Temperature int     //(min -40 °C max 215)
}

//Emulator тип для эмуляции датчика ДУТ-ЭМ
type Emulator struct {
	Enabled bool
	Sensors [numSensors]FuelParams
}

//Enable - включить эмуляцию датчика с номером sensorNumber (от 0 до 7)
func (dut *Emulator) Enable(sensorNumber uint) {
	if sensorNumber < numSensors {
		dut.Sensors[sensorNumber].Enabled = true
	}
}

//Disable - выключить эмуляцию датчика с номером sensorNumber (от 0 до 7)
func (dut *Emulator) Disable(sensorNumber uint) {
	if sensorNumber < numSensors {
		dut.Sensors[sensorNumber].Enabled = false
	}
}

//SetLevel установить уровень для датчика с номером sensorNumber (от 0 до 7).
//level - минимальное значение 0.0 м, максимальное 1.0 м
func (dut *Emulator) SetLevel(sensorNumber uint, level float64) {
	if dut != nil && sensorNumber < numSensors {
		dut.Sensors[sensorNumber].Level = level
	}
}

//SetTemperature установить температуру для датчика с номером sensorNumber (от 0 до 7).
//temperature - минимальное значение -40 °С
func (dut *Emulator) SetTemperature(sensorNumber uint, temperature int) {
	if dut != nil && sensorNumber < numSensors {
		dut.Sensors[sensorNumber].Temperature = temperature
	}
}

//Set установить параметры для датчика с номером sensorNumber (от 0 до 7).
//level - минимальное значение 0.0 м, максимальное 1.0 м
//temperature - минимальное значение -40 °С
func (dut *Emulator) Set(sensorNumber uint, level float64, temperature int) {
	if dut != nil {
		dut.SetLevel(sensorNumber, level)
		dut.SetTemperature(sensorNumber, temperature)
	}
}

//конвертирует заданный уровень для вставки в сообщение CAN
func convertLevel(level float64) uint16 {
	lvl := level
	if lvl < 0. {
		lvl = 0.
	}
	if lvl > 1.0 {
		lvl = 1.0
	}
	return uint16(10000 * lvl)
}

//конвертирует заданную температуру для вставки в сообщение CAN
func convertTemp(temper int) byte {
	iTemp := temper + 40
	if iTemp < 0 {
		iTemp = 0
	}
	if iTemp > 255 {
		iTemp = 255
	}
	return byte(uint(iTemp))
}

//Start запускает эмуляцию датчика ДУТ-ЭМ.
//dev - устройство CAN, в которое нужно выдавать сообщения датчика.
//Функция возвращает указатель, к котому нужно обращаться в дальнейшем (вызывать его методы)
func (dut *Emulator) Start(dev *candev.Device) {
	if nil == dev {
		panic("Попытка запустить эмуляцию ДУТ-ЭМ с устройством CAN == nil")
	}

	if !dut.Enabled {
		go func() {
			fmt.Println("Запуск эмуляции датчиков ДУТ-ЭМ (CAN 250)")
			dut.Enabled = true
			msg := candev.Message{Len: 8}
			for dut.Enabled {
				for i, sens := range dut.Sensors {
					if sens.Enabled {
						msg.ID = baseDUT + uint32(i+1)

						binary.LittleEndian.PutUint16(msg.Data[0:], convertLevel(sens.Level))
						msg.Data[6] = convertTemp(sens.Temperature)

						dev.Send(msg)

						// fmt.Printf("%X %X\r\n", msg.ID, msg.Data)
						fmt.Println("DUT")
						time.Sleep(time.Second)
					}
				}
			}
		}()
	}
}

//Stop останавливает запущенную эмуляцию.
func (dut *Emulator) Stop() {
	dut.Enabled = false
}
