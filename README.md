# Go Phone
Private, do not use!

## Kickstart

### Do it once
- `cp .env.dict .env` ...and customize
### Some make
- `make init` to init PC environment
- `make go-build` to compile code
- `make arm-authorize` to copy ssh key to ARM (login without password)
- `make arm-install` to copy binary file to ARM

### If you need RPI console (ssh)

- `make rpi-console`

## Notes
In `/usr/share/alsa/alsa.conf`
Change to 1 or 2

```
defaults.ctl.card 1
defaults.pcm.card 1
defaults.pcm.device 0
defaults.pcm.subdevice 0
```

`speaker-test`
