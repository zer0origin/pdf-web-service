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

    constructor() {
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
     * Determines the top-left corner by subtracting half the width and height from the center point. It finds the
     * bottom-right corner by adding half the width and height to that same center point
     * @param p1 {Point}
     * @param p2 {Point}
     */
    function getPointsAsUpperLeftAndLowerRight(p1, p2) {
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

    /**
     * Spawn a new point on the screen
     * @param p {Point}
     * @param element {HTMLElement}
     */
    function spawnPoint(p, element){
        let temp = document.getElementsByTagName("template")[0]
        let node = document.importNode(temp.content.querySelector("div"))
        element.appendChild(node)
    }

    function spawnSelection(){

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
        console.log(imageCoordsRelativeToSelf)

        if (imageCoordsRelativeToSelf.x < 0){
            imageCoordsRelativeToSelf.x = 0
        }

        if (imageCoordsRelativeToSelf.y < 0){
            imageCoordsRelativeToSelf.y = 0
        }

        let name = String(event.target.id);
        let key = Number(name.split("-")[1]); //todo: change
        let selectionArr = `selection-${key}`
        console.log(key)

        let present = selectionsMap.has(key)
        if (!present){
            console.log("NEW ARRAY")
            selectionsMap.set(key, [])
        }else{
            console.log(selectionsMap.get(key)) //debug
        }

        let recArr = selectionsMap.get(key);

        if (recArr.length <= 0){
            let rec = new Rectangle();
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            return
        }

        let recData = recArr[recArr.length - 1];
        if (recData.p2 === undefined){
            recData.p2 = imageCoordsRelativeToSelf;
            return;
        }else{
            let rec = new Rectangle();
            rec.p1 = imageCoordsRelativeToSelf;
            recArr.push(rec)
            return;
        }
    }

    return {
        getPosition: getPositionRelativeToDocument,
        onClick: onClickFunction,
        selectionsMap: selectionsMap,
        spawnPoint: spawnPoint,
    };
})()