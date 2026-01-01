module payment-service

go 1.25.3


//replace tiny-p2p => ../tiny-lib
//require tiny-p2p v0.0.0-00010101000000-000000000000
replace github.com/georghagn/gsf-suite => ../../../..
require github.com/georghagn/gsf-suite v0.0.0-00010101000000-000000000000
require github.com/coder/websocket v1.8.14 // indirect
