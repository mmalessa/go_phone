# Go Phone
Private, do not use!

## Kickstart

### Prepare Orengepi Zero
`armbian-config` -> System -> Hardware -> [*] analog codec
`armbian-config` -> Network -> Hotspot -> wlan0

### Do it once
- `cp .env.dict .env` ...and customize
- 
### Some make
- `make init` to init PC environment
- `make go-build` to compile code
- `make arm-authorize` to copy ssh key to ARM (login without password)
- `make arm-init` to "init" some stuff in ARM
- `make arm-install` to copy binary file to ARM
- `make rpi-console`


## Some notes



### dts
```
linux,code = <79>; /* KEY_KP1, see /usr/include/linux/input-event-codes.h */
gpios = <&pio 4 11 0>; /* PE11 GPIO_ACTIVE_HIGH */
```

`apt install input-utils`
```sh
lsinput
/dev/input/event0
   bustype : BUS_HOST
   vendor  : 0x1
   product : 0x1
   version : 256
   name    : "gpio-keys-user"
   phys    : "gpio-keys/input0"
   bits ev : (null) (null)

input-events 0

```