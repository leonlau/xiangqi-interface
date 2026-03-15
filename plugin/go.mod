module github.com/leonlau/xiangqi-interface/plugin

go 1.26

require (
	github.com/leonlau/xiangqi v0.0.0
	github.com/leonlau/xiangqi-interface v0.0.0
)

require github.com/ajstarks/svgo v0.0.0-20200320125537-f189e35d30ca // indirect

replace github.com/leonlau/xiangqi-interface => ../

replace github.com/leonlau/xiangqi => ../vendor/engine
