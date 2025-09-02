class Point{
    x;
    y;
    constructor(x, y) {
        this.x = x;
        this.y = y;
    }
}

class Rectangle{
    p1;
    p2;
    spawnDiv;
    position;

    /**
     * @type {[HTMLElement]}
     */
    nodes = [];

    /**
     * @type {HTMLElement}
     */
    rectangleDiv = undefined

    constructor(spawnDiv, position) {
        this.spawnDiv = spawnDiv;
        this.position = position;
    }

    /**
     * Spawn a new point on the screen
     * @param p {Point}
     * @param element {HTMLElement}
     * @param position {{left: number, top: number}}
     */
    #spawnPoint(p, element, position){
        let size = 5;
        let temp = document.getElementById("point-template")
        let node = document.importNode(temp.content.querySelector("div"))
        element.appendChild(node)
        this.nodes.push(node)
        console.log(this.nodes)

        node.style.width = `${size}px`;
        node.style.height = `${size}px`;
        node.style.top = `${(position.top + p.y) - (size/2)}px`
        node.style.left = `${(position.left + p.x) - (size/2)}px`
    }

    /**
     * Spawn a new point on the screen
     */
    spawnP1(){
        this.#spawnPoint(this.p1, this.spawnDiv, this.position)
    }

    /**
     * Spawn a new point on the screen
     */
    spawnP2(){
        this.#spawnPoint(this.p2, this.spawnDiv, this.position)
    }

    /**
     * Remove all the spawned points.
     */
    clearSpawnedPoints(){
        while (this.nodes.length > 0){
            let element = this.nodes.pop();
            element.remove();
        }
    }

    /**
     * Spawn a new point on the screen
     * @param spawnDiv {HTMLElement}
     * @param position {{left: number, top: number}}
     */
    spawnRectangle(spawnDiv, position){
        let temp = document.getElementById("rectangle-template")
        let node = document.importNode(temp.content.querySelector("div"))
        spawnDiv.appendChild(node)
        this.rectangleDiv = node
        console.log(this.rectangleDiv)

        let points = this.#getPointsAsUpperLeftAndLowerRight(this.p1, this.p2)
        let width = points.p2.x - points.p1.x
        let height = points.p2.y - points.p1.y

        node.style.width = `${width}px`;
        node.style.height = `${height}px`;
        node.style.top = `${position.top + points.p1.y}px`
        node.style.left = `${position.left + points.p1.x}px`
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


        let rec = new Rectangle();
        rec.p1 = upperLeft;
        rec.p2 = lowerRight;

        return rec
    }
}

const selectionsModule = (function() {
    /**
     *
     * @type {Map<number, Array<Rectangle>>}
     */
    let selectionsMap = new Map();

    /**
     *
     * @param {HTMLElement} element
     * @returns {{left: number, top: number}}
     */
    function getPositionRelativeToDocument(element) {
        const clientRect = element.getBoundingClientRect();
        return {
            left: clientRect.left + window.pageXOffset,
            top: clientRect.top + window.pageYOffset
        };
    }

    /**
     * Handle click events for selections.
     * @param {MouseEvent} event - The mouse event object.
     */
    function onClickFunction(event) {
        const cursorX = event.pageX; //get cursorX relative to whole page.
        const cursorY = event.pageY; //get cursorY relative to whole page.

        let imagePos = getPositionRelativeToDocument(event.target)
        let zoomLevel = zoomModule.getZoomLevel();
        const imageCoordsRelativeToSelf = new Point((cursorX - imagePos.left) / zoomLevel, (cursorY - imagePos.top) / zoomLevel)

        if (imageCoordsRelativeToSelf.x < 0){
            imageCoordsRelativeToSelf.x = 0
        }

        if (imageCoordsRelativeToSelf.y < 0){
            imageCoordsRelativeToSelf.y = 0
        }

        let name = String(event.target.id);
        let key = Number(name.split("-")[1]); //todo: change
        let selectionArr = `selection-${key}`

        let present = selectionsMap.has(key)
        if (!present){
            selectionsMap.set(key, [])
        }

        let recArr = selectionsMap.get(key);

        if (recArr.length <= 0){
            let rec = new Rectangle(document.getElementById(selectionArr), getPositionRelativeToDocument(event.target));
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            rec.spawnP1()
            return
        }

        let recData = recArr[recArr.length - 1];
        if (recData.p2 === undefined){
            recData.p2 = imageCoordsRelativeToSelf;
            recData.spawnP2()
            // recData.clearSpawnedPoints();
            recData.spawnRectangle(document.getElementById(selectionArr), getPositionRelativeToDocument(event.target))
            return;
        }else{
            let rec = new Rectangle(document.getElementById(selectionArr), getPositionRelativeToDocument(event.target));
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            rec.spawnP1()
            return;
        }
    }

    return {
        getPosition: getPositionRelativeToDocument,
        onClick: onClickFunction,
        selectionsMap: selectionsMap,
    };
})()