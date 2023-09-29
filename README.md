# ValetudoPNG

Valetudo map renderer (alternative to [Hypfer/ICantBelieveItsNotValetudo](https://github.com/Hypfer/ICantBelieveItsNotValetudo)), written in Go.

# Motyvation

The [Hypfer/ICantBelieveItsNotValetudo](https://github.com/Hypfer/ICantBelieveItsNotValetudo) project has quite few nuances:

1. It's literally not working (see [here](https://github.com/Hypfer/ICantBelieveItsNotValetudo/pull/92)).
2. Uncomfortably long name.
3. Author does not create CI/CD job for Docker image building (and _deleted_ my raised issue where I literally provided the full code as a suggestion).
4. Has no instructions, guidelines or output for calibration points to be used with [PiotrMachowski/lovelace-xiaomi-vacuum-map-card](https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card).
5. Written in Javascript for PNG rendering. This is not the right language for resource-intensive task.
6. Image cropping does not exist.

Okay, there is also [rand256/valetudo-mapper](https://github.com/rand256/valetudo-mapper) which also crashes with Valetudo and basically does not work at all...

# Features

* Written in Go.
  * Single binary
  * No dependencies
  * Fast & multithreaded rendering
* Pre-built Docker images.
* Automatic map calibration data for [PiotrMachowski/lovelace-xiaomi-vacuum-map-card](https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card).
* Easy configuration using `yaml` config file.
* Map modification:
  * Rotation
  * Scaling
  * "croping" by binding map to coordinates in robot's coordinates system
* HTTP endpoint:
  * Access image `http://ip:port/api/map/image`.
  * Debug image and it's coordinates/pixels in robot's coordinates system `http://ip:port/api/map/image/debug`.
  * Designed to work with HomeAssistant in mind.

Supported architectures:

* linux/amd64
* linux/arm64
* linux/armv7
* linux/armv6

# Get started

## Configure Valetudo

It is assumed that Valetudo is connected to Home Assistant via MQTT and is working.

Go to Valetudo URL -> Connectivity -> MQTT Connectivity -> Customizations. Make sure `Provide map data` is enabled.

## Configuration file

Create `config.yml` file out of `config.example.yml` file and update according.

For starters, assuming that you don't have TLS and username/password set in your MQTT server, you can update only these for now:
```yaml
    host: 192.168.0.123
    port: 1883
```
and these:
```yaml
    # Should match "Topic prefix" in Valetudo MQTT settings
    valetudo_prefix: valetudo

    # Should match "Identifier" in Valetudo MQTT settings
    valetudo_identifier: rockrobo
```

Now move to installation and usage sections, where you will be able to easily "experiment" with your config.

## Installation

### Binaries

See [Releases](https://github.com/erkexzcx/valetudopng/releases).

```bash
$ tar -xvzf valetudopng_v1.0.0_linux_amd64.tar.gz 
valetudopng_v1.0.0_linux_amd64
$ ./valetudopng_v1.0.0_linux_amd64 --help
Usage of ./valetudopng_v1.0.0_linux_amd64:
  -config string
        Path to configuration file (default "config.yml")
  -version
        prints version of the application
```

### Docker compose

```yaml
  valetudopng:
    image: ghcr.io/erkexzcx/valetudopng:latest
    container_name: valetudopng
    restart: always
    volumes:
      - ./valetudopng/config.yml:/config.yml
    ports:
      - "3000:3000"
```

### Docker CLI

```
docker run -d \
    --restart=always \
    --name=valetudopng \
    -v $(pwd)/valetudopng/config.yml:/config.yml \
    -p 3000:3000 \
    ghcr.io/erkexzcx/valetudopng:latest
```

## Usage

When hosted, go to `http://ip:port/api/map/image/debug` and start selecting rectangles. Below the picture there will be information that you will want to copy/paste.

For example, this is how my [PiotrMachowski/lovelace-xiaomi-vacuum-map-card](https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card) card looks like with `valetudo_prefix: valetudo` and `valetudo_identifier: rockrobo` and RockRobo S5 vacuum:

```yaml
type: custom:xiaomi-vacuum-map-card
map_source:
  camera: camera.rockrobo_rendered_map
calibration_source:
  entity: sensor.rockrobo_calibration_data
entity: vacuum.valetudo_rockrobo
vacuum_platform: Hypfer/Valetudo
internal_variables:
  topic: valetudo/rockrobo
map_modes:
  - template: vacuum_clean_zone_predefined
    predefined_selections:
      - id: Entrance
        outline: [[2310,2975],[2185,2975],[2185,3090],[2310,3090]]
        label:
          text: Entrance
          x: 2247.5
          y: 3032.5
          offset_y: 28
        icon:
          name: mdi:door
          x: 2247.5
          y: 3032.5
  - template: vacuum_goto
  - template: vacuum_clean_zone
map_locked: true
two_finger_pan: false
```
