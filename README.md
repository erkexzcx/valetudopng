# ValetudoPNG

ValetudoPNG is a service designed to render map from Valetudo-enabled vacuum robot into a more accessible PNG format. This PNG map is sent to Home Assistant via MQTT, where it can be viewed as a real-time camera feed. ValetudoPNG was specifically developed to integrate with third-party frontend cards, such as the [PiotrMachowski/lovelace-xiaomi-vacuum-map-card](https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card).

Alternative projects:
* [sca075/valetudo_vacuum_camera](https://github.com/sca075/valetudo_vacuum_camera) - deploys as HACS addon, written in Python.

Broken or dead projects:
* [Hypfer/ICantBelieveItsNotValetudo](https://github.com/Hypfer/ICantBelieveItsNotValetudo) - original project, written in javascript for NodeJS.
* [rand256/valetudo-mapper](https://github.com/rand256/valetudo-mapper) - fork of original project to be used with [rand256/valetudo](https://github.com/rand256/valetudo). Does not work with [Hypfer/Valetudo](https://github.com/Hypfer/Valetudo).

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

You can technically install it on robot itself:
```
[root@rockrobo ~]# grep -e scale config.yml
  scale: 2
[root@rockrobo ~]# ./valetudopng
2023/10/02 07:18:10 [MQTT producer] Connected
2023/10/02 07:18:10 [MQTT consumer] Connected
2023/10/02 07:18:10 [MQTT consumer] Subscribed to map data topic
2023/10/02 07:18:10 Image rendered! drawing:39ms, encoding:61ms, size:9.1kB
2023/10/02 07:18:16 Image rendered! drawing:37ms, encoding:72ms, size:9.1kB
2023/10/02 07:18:16 Image rendered! drawing:35ms, encoding:66ms, size:9.1kB
2023/10/02 07:18:17 Image rendered! drawing:44ms, encoding:50ms, size:7.4kB
2023/10/02 07:18:18 Image rendered! drawing:33ms, encoding:54ms, size:7.4kB
2023/10/02 07:18:20 Image rendered! drawing:34ms, encoding:52ms, size:7.4kB
2023/10/02 07:18:22 Image rendered! drawing:34ms, encoding:61ms, size:7.4kB
2023/10/02 07:18:24 Image rendered! drawing:32ms, encoding:56ms, size:7.7kB
2023/10/02 07:18:26 Image rendered! drawing:45ms, encoding:62ms, size:7.8kB
2023/10/02 07:18:28 Image rendered! drawing:33ms, encoding:64ms, size:7.8kB
2023/10/02 07:18:30 Image rendered! drawing:44ms, encoding:59ms, size:8.0kB
2023/10/02 07:18:32 Image rendered! drawing:38ms, encoding:62ms, size:8.2kB
2023/10/02 07:18:35 Image rendered! drawing:88ms, encoding:54ms, size:8.3kB
2023/10/02 07:18:36 Image rendered! drawing:35ms, encoding:72ms, size:8.4kB
```
Download binary appropriate for your robot's CPU and follow the service installation guidelines of another project: https://github.com/porech/roborock-oucher

Note that this service is still resources-intensive and drains more battery when robot is not charging. Generally it is not recommended to host it on robot.

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
    selection_type: PREDEFINED_RECTANGLE
    predefined_selections:
      - zones: [[2185,2975,2310,3090]]
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
