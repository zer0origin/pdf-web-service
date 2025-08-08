let isEnabled = false;

function selections(event) {
    isEnabled = true;
}

document.ad

/**
 * Handles click events.
 * @param {MouseEvent} event - The mouse event object.
 */
function onClickFunction(event) {
    console.log("Fired")
    if (!isEnabled) {
        return;
    }

    const cursorX = event.pageX; //get cursorX relative to whole page.
    const cursorY = event.pageY; //get cursorY relative to whole page.

    console.log(cursorX)
    console.log(cursorY)
    console.log(event.target)

    imagePos = getPositionRelativeToDocument(event.target)
    zoomLevel = 1;
    const a = {"x": (cursorX - imagePos.left) / zoomLevel, "y": (cursorY - imagePos.top) / zoomLevel}

    console.log(a)
}

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