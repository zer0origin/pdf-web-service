Array.prototype.remove = function (from, to) {
    var rest = this.slice((to || from) + 1 || this.length);
    this.length = from < 0 ? this.length + from : from;
    return this.push.apply(this, rest);
};

class Point {
    x;
    y;

    constructor(x, y) {
        this.x = x;
        this.y = y;
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
    p1;
    p2;
    spawnDiv;
    imageDiv;
    id;

    /**
     * @type {[HTMLElement]}
     */
    nodes = [];

    /**
     * @type {HTMLElement}
     */
    rectangleDiv = undefined

    constructor(spawnDiv, imageDiv) {
        this.spawnDiv = spawnDiv;
        this.imageDiv = imageDiv;
        console.log("BEFORE: " + Rectangle.lastId)
        this.id = Rectangle.lastId++;
        console.log("AFTER: " + Rectangle.lastId)
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
        let node = document.importNode(temp.content.querySelector("div"), true)
        element.appendChild(node)
        this.nodes.push(node)

        node.style.width = `${size}px`;
        node.style.height = `${size}px`;
        node.style.top = `${(position.top + (p.y * zoomModule.getZoomLevel())) - (size / 2)}px`
        node.style.left = `${(position.left + (p.x * zoomModule.getZoomLevel())) - (size / 2)}px`
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
     * Remove all the spawned points.
     */
    clearSpawnedPoints() {
        while (this.nodes.length > 0) {
            let element = this.nodes.pop();
            element.remove();
        }
    }

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
        let height = (points.p2.y - points.p1.y)  * zoomModule.getZoomLevel()

        node.style.width = `${width}px`;
        node.style.height = `${height}px`;
        let position = Point.getPositionRelativeToDocument(this.imageDiv);
        node.style.top = `${position.top + (points.p1.y * zoomModule.getZoomLevel())}px`
        node.style.left = `${position.left + (points.p1.x * zoomModule.getZoomLevel())}px`

        let exitControls = node.querySelector(".exit-controls")

        exitControls.onclick = () => {
            let name = String(this.imageDiv.id);
            let key = Number(name.split("-")[1]); //todo: change
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

const selectionsModule = (function () {
    /**
     *
     * @type {Map<number, Array<Rectangle>>}
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
        let pageNumber = Number(name.split("-")[1]); //todo: change
        let selectionArr = `selection-${pageNumber}`

        let present = selectionsMap.has(pageNumber)
        if (!present) {
            selectionsMap.set(pageNumber, [])
        }

        let recArr = selectionsMap.get(pageNumber);
        if (recArr.length <= 0) {
            let rec = new Rectangle(document.getElementById(selectionArr), event.target);
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
            let rec = new Rectangle(document.getElementById(selectionArr), event.target);
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            rec.spawnP1()
        }
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

    function refreshSelectionNodes() {
        for (let i = 0; i < selectionsMap.size; i++) {
            let recArr = selectionsMap.get(i);

            for (let j = 0; j < recArr.length; j++) {
                let rec = recArr[j];
                rec.clearSpawnedNodes();
                rec.spawnRectangle()
            }
        }

        console.log("refreshed viewer!")
    }

    return {
        map: selectionsMap,
        onClick: onClickFunction,
        deleteSelection: deleteSelection,
        refreshSelectionNodes: refreshSelectionNodes,
    };
})()

window.addEventListener("resize", selectionsModule.refreshSelectionNodes);
zoomModule.registerZoomChange(selectionsModule.refreshSelectionNodes);
