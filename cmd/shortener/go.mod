module main

go 1.19

replace internal/app => ../../internal/app

replace internal/handler => ../../internal/handler

require internal/app v0.0.0-00010101000000-000000000000

require internal/handler v0.0.0-00010101000000-000000000000 // indirect
