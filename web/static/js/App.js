function calculateMapPosition(imageX, imageY) {
    // Calculate the robot's coordinates
    var robotX = Math.floor(imageX / scale);
    var robotY = Math.floor(imageY / scale);

    // Adjust the map's coordinates based on the rotation
    var rotatedRobotX, rotatedRobotY;
    switch (rotatedTimes) {
        case 0:
            // No rotation
            rotatedRobotX = robotMinX + robotX;
            rotatedRobotY = robotMinY + robotY;
            break;
        case 1:
            // 90 degrees clockwise
            rotatedRobotX = robotMinX + robotY;
            rotatedRobotY = robotMaxY - robotX;
            break;
        case 2:
            // 180 degrees clockwise
            rotatedRobotX = robotMaxX - robotX;
            rotatedRobotY = robotMaxY - robotY;
            break;
        case 3:
            // 270 degrees clockwise
            rotatedRobotX = robotMaxX - robotY;
            rotatedRobotY = robotMinY + robotX;
            break;
    }

    imageX = imageX - (imageX%scale)
    imageY = imageY - (imageY%scale)

    return {
        mapStartX: imageX, mapStartY: imageY,
        mapEndX: imageX+scale, mapEndY: imageY+scale,
        robotX: rotatedRobotX*pixelSize, robotY: rotatedRobotY*pixelSize
    };
}

// Create "Canvas drawing object" (just the way I call it lol)
let cdo = {};
cdo.rectCoords = [];

cdo.clear = function() {
    this.ctx.clearRect(0, 0, this.canvas.width(), this.canvas.height());
}

cdo.drawCrosshair = function (rawImgX, rawImgY) {
    var pos = calculateMapPosition(rawImgX, rawImgY);

    // Draw highlight for pixel block
    this.ctx.fillStyle = 'rgba(255, 0, 0, 1)';
    this.ctx.fillRect(pos.mapStartX, pos.mapStartY, pos.mapEndX - pos.mapStartX, pos.mapEndY - pos.mapStartY);

    // Draw highlights for horizontal and vertical axis rows
    this.ctx.fillStyle = 'rgba(255, 0, 0, 0.25)';
    this.ctx.fillRect(pos.mapStartX, 0, pos.mapEndX - pos.mapStartX, this.canvas.height());
    this.ctx.fillRect(0, pos.mapStartY, this.canvas.width(), pos.mapEndY - pos.mapStartY);
}

cdo.drawRectangle = function (rawImgX, rawImgY) {
    if (cdo.rectCoords.length == 0 || cdo.rectCoords.length > 2) {
        return
    }

    sx = cdo.rectCoords[0][0];
    sy = cdo.rectCoords[0][1];

    if (cdo.rectCoords.length == 1) {
        var position = calculateMapPosition(rawImgX, rawImgY);
        ex = position.mapStartX+scale;
        ey = position.mapStartY+scale;
    } else {
        ex = cdo.rectCoords[1][0]+scale;
        ey = cdo.rectCoords[1][1]+scale;
    }

    if (sx > ex || sy > ey) {
        let temp = sx;
        sx = ex;
        ex = temp;
    
        temp = sy;
        sy = ey;
        ey = temp;
    }

    width = ex-sx;
    height = ey-sy
    
    this.ctx.fillStyle = 'rgba(158, 221, 255, 0.60)';
    this.ctx.fillRect(sx, sy, width, height);

    res1 = calculateMapPosition(sx, sy);
    res2 = calculateMapPosition(ex-scale, ey-scale);

    cdo.setCardData(res1.robotX, res1.robotY, res2.robotX, res2.robotY);
    cdo.setCofigData(res1.robotX, res1.robotY, res2.robotX, res2.robotY);
}

cdo.addRectangleCoordinates = function (rawImgX, rawImgY) {
    if (cdo.rectCoords.length < 2) {
        var position = calculateMapPosition(rawImgX, rawImgY);
        cdo.rectCoords.push([position.mapStartX, position.mapStartY]);
        return;
    }else{
        cdo.rectCoords = [];
        return;
    }
}

cdo.showPopup = function (rawImgX, rawImgY, pageX, pageY) {
    var position = calculateMapPosition(rawImgX, rawImgY);
    this.popup.text('Map: (' + position.mapStartX + ', ' + position.mapStartY + '), Robot: (' + position.robotX + ', ' + position.robotY + ')')
        .css({left: pageX + 10, top: pageY + 10})
        .show();
}

cdo.hidePopup = function () {
    this.popup.hide();
}

// Set robot coordinates, not map coordinates
// for xiaomi-vacuum-map-card
cdo.setCardData = function(x1, y1, x2, y2){
    middleX = (x2+x1)/2
    middleY = (y2+y1)/2
    str = `map_modes:
  - template: vacuum_clean_zone_predefined
    # See https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card/issues/662
    selection_type: PREDEFINED_RECTANGLE
    predefined_selections:
      - zones: [[${x1},${y1},${x2},${y2}]]
        label:
          text: Entrance
          x: ${middleX}
          y: ${middleY}
          offset_y: 28
        icon:
          name: mdi:door
          x: ${middleX}
          y: ${middleY}`;
    cdo.carddata.html(str);
}

// Set robot coordinates, not map coordinates
// for config file (YAML)
cdo.setCofigData = function(x1, y1, x2, y2){
    str = `
  custom_limits:
    start_x: ${x1}
    start_y: ${y1}
    end_x: ${x2}
    end_y: ${y2}`;
    cdo.cofigdata.html(str);
}
