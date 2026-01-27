class Point {
    x;
    y;

    constructor(x, y) {
        this.x = x;
        this.y = y;
    }

    copy() {
        return new Point(this.x, this.y)
    }

    static getPositionRelativeToDocument(element) {
        const clientRect = element.getBoundingClientRect();
        return {
            left: clientRect.left + window.pageXOffset,
            top: clientRect.top + window.pageYOffset
        };
    }
}

class Rectangle {
    static lastId = 0
    /**
     * @type {Point}
     */
    p1;
    /**
     * @type {Point}
     */
    p2;
    /**
     * @type {HTMLDivElement}
     */
    spawnDiv;
    /**
     * @type {HTMLDivElement}
     */
    imageDiv;
    id;

    /**
     * @type {[HTMLElement]}
     */
    pointNodesArr = [];
    /**
     * @type {HTMLElement}
     */
    rectangleDiv = undefined

    constructor(spawnDiv, imageDiv, p1 = undefined, p2 = undefined, id = undefined) {
        this.spawnDiv = spawnDiv;
        this.imageDiv = imageDiv;
        this.p1 = p1;
        this.p2 = p2;

        if (id !== undefined) {
            this.id = id;
            return;
        }

        this.id = Rectangle.lastId++;
    }

    /**
     * Spawn a new point on the screen
     * @param p {Point}
     * @param element {HTMLElement}
     * @param position {{left: number, top: number}}
     */
    #spawnPoint(p, element, position) {
        let size = 5;
        let temp = document.getElementById("point-template")
        let pointDiv = document.importNode(temp.content.querySelector("div"), true)
        element.appendChild(pointDiv)
        this.pointNodesArr.push(pointDiv)

        pointDiv.style.width = `${size}px`;
        pointDiv.style.height = `${size}px`;
        pointDiv.style.top = `${(position.top + (p.y * zoomModule.getZoomLevel())) - (size / 2)}px`
        pointDiv.style.left = `${(position.left + (p.x * zoomModule.getZoomLevel())) - (size / 2)}px`
    }

    /**
     * Spawn a new point on the screen
     */
    spawnP1() {
        this.#spawnPoint(this.p1, this.spawnDiv, Point.getPositionRelativeToDocument(this.imageDiv));
    }

    /**
     * Spawn a new point on the screen
     */
    spawnP2() {
        this.#spawnPoint(this.p2, this.spawnDiv, Point.getPositionRelativeToDocument(this.imageDiv));
    }

    /**
     * Remove all the spawned points. Points are visual elements on the screen, that have yet to become rectangles.
     */
    clearSpawnedPoints() {
        while (this.pointNodesArr.length > 0) {
            let element = this.pointNodesArr.pop();
            element.remove();
        }
    }

    /**
     * Remove spawned rectangles.
     */
    clearSpawnedRectangle() {
        if (this.rectangleDiv) {
            this.rectangleDiv.remove()
        }
    }

    /**
     * Spawn a new point on the screen
     */
    spawnRectangle() {
        let temp = document.getElementById("rectangle-template")
        let node = document.importNode(temp.content.querySelector("div"), true)
        this.rectangleDiv = node

        let points = this.#getPointsAsUpperLeftAndLowerRight(this.p1, this.p2)
        let width = (points.p2.x - points.p1.x) * zoomModule.getZoomLevel()
        let height = (points.p2.y - points.p1.y) * zoomModule.getZoomLevel()

        node.style.width = `${width}px`;
        node.style.height = `${height}px`;
        let position = Point.getPositionRelativeToDocument(this.imageDiv);
        node.style.top = `${position.top + (points.p1.y * zoomModule.getZoomLevel())}px`
        node.style.left = `${position.left + (points.p1.x * zoomModule.getZoomLevel())}px`

        let exitControls = node.querySelector(".exit-controls")

        exitControls.onclick = () => {
            let name = String(this.imageDiv.id);
            let key = name.split("-")[1];
            selectionsModule.deleteSelection(key, this.id)
        }

        this.spawnDiv.appendChild(node)
    }

    /**
     * Determines the top-left corner by subtracting half the width and height from the center point. It finds the
     * bottom-right corner by adding half the width and height to that same center point
     * @param p1 {Point}
     * @param p2 {Point}
     */
    #getPointsAsUpperLeftAndLowerRight(p1, p2) {
        let midX = (p1.x + p2.x) / 2;
        let midY = (p1.y + p2.y) / 2;
        let diffX = Math.abs(p1.x - p2.x) / 2;
        let diffY = Math.abs(p1.y - p2.y) / 2

        let upperLeft = new Point(midX - diffX, midY - diffY)
        let lowerRight = new Point(midX + diffX, midY + diffY)

        return {p1: upperLeft, p2: lowerRight}
    }

    clearSpawnedNodes() {
        this.clearSpawnedPoints();
        this.clearSpawnedRectangle();
    }
}

var selectionsModule = (function () {
    Array.prototype.remove = function (from, to) {
        var rest = this.slice((to || from) + 1 || this.length);
        this.length = from < 0 ? this.length + from : from;
        return this.push.apply(this, rest);
    };

    /**
     *
     * @type {Map<string, Array<Rectangle>>}
     */
    let selectionsMap = new Map();

    /**
     * Handle click events for selections.
     * @param {MouseEvent} event - The mouse event object.
     */
    function onClickFunction(event) {
        const cursorX = event.pageX; //get cursorX relative to whole page.
        const cursorY = event.pageY; //get cursorY relative to whole page.

        let imagePos = Point.getPositionRelativeToDocument(event.target)
        let zoomLevel = zoomModule.getZoomLevel();
        const imageCoordsRelativeToSelf = new Point((cursorX - imagePos.left) / zoomLevel, (cursorY - imagePos.top) / zoomLevel)

        if (imageCoordsRelativeToSelf.x < 0) {
            imageCoordsRelativeToSelf.x = 0
        }

        if (imageCoordsRelativeToSelf.y < 0) {
            imageCoordsRelativeToSelf.y = 0
        }

        let name = String(event.target.id);
        let pageKey = name.split("-")[1];
        let spawnDiv = `selection-${pageKey}`

        let present = selectionsMap.has(pageKey)
        if (!present) {
            selectionsMap.set(pageKey, [])
        }

        let recArr = selectionsMap.get(pageKey);
        if (recArr.length <= 0) {
            let rec = new Rectangle(document.getElementById(spawnDiv), event.target);
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            rec.spawnP1()
            return
        }

        let recData = recArr[recArr.length - 1];
        if (recData.p2 === undefined) {
            recData.p2 = imageCoordsRelativeToSelf;
            recData.spawnP2()
            recData.clearSpawnedPoints();
            recData.spawnRectangle()
        } else {
            let rec = new Rectangle(document.getElementById(spawnDiv), event.target);
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            rec.spawnP1()
        }
    }

    /**
     * @param pageKey {string}
     * @param rec {Rectangle}
     */
    function pushSelectionToMap(pageKey, rec) {
        let present = selectionsMap.has(pageKey)
        if (!present) {
            selectionsMap.set(pageKey, [])
        }

        let recArr = selectionsMap.get(pageKey);
        recArr.push(rec);
    }

    /**
     * Spawns a rectangle on an image.
     * @param pageKey {string}
     * @param p1 {Point}
     * @param p2 {Point}
     */
    function loadRectangle(pageKey, p1, p2, id = undefined) {
        let spawnDiv = document.getElementById(`selection-${pageKey}`);
        let imageDiv = document.getElementById(`image-${pageKey}`);

        let rec = new Rectangle(spawnDiv, imageDiv, p1, p2, id);
        pushSelectionToMap(pageKey, rec);
        rec.spawnRectangle();
    }

    function deleteSelection(key, id) {
        let rectangles = selectionsMap.get(key);

        for (let i = 0; i < rectangles.length; i++) {
            let rec = rectangles[i];
            if (rec.id === id) {
                rec.clearSpawnedNodes()
                rectangles.remove(i)
            }
        }
    }

    /**
     * Redraw selection nodes.
     */
    function redrawSelectionNodes() {
        if (selectionsMap.size <= 0) {
            return
        }

        selectionsMap.forEach((recArr, i) => {
            recArr.forEach((rec, j) => {
                rec.clearSpawnedNodes();
                rec.spawnRectangle()
            })
        })
    }

    return {
        map: selectionsMap, //For Debug
        load: loadRectangle,
        onClick: onClickFunction,
        deleteSelection: deleteSelection,
        refreshSelectionNodes: redrawSelectionNodes,
    };
})()

window.addEventListener("resize", selectionsModule.refreshSelectionNodes);

if (window.zoomModule) {
    zoomModule.registerZoomChange(selectionsModule.refreshSelectionNodes);
} else {
    window.addEventListener("zoomModuleLoaded", () => {
        console.log("Event Fired!")
        zoomModule.registerZoomChange(selectionsModule.refreshSelectionNodes);
    })
}
