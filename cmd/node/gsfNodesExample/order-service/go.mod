module order-service

go 1.25.3

//require tiny-p2p v0.0.0
//replace tiny-p2p => ../tiny-lib
require github.com/georghagn/gsf-suite v0.0.0-00010101000000-000000000000
replace github.com/georghagn/gsf-suite => ../../../..

require github.com/coder/websocket v1.8.14 // indirect