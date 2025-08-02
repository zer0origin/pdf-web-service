let isEnabled = false;

function selections() {
    isEnabled = true;
}

function onClickFunction(event) {
    console.log("Fired")
    if (!isEnabled) {
        return;
    }

    const cursorX = event.pageX; //get cursorX relative to whole page.
    const cursorY = event.pageY; //get cursorY relative to whole page.

    console.log(cursorX)
    console.log(cursorY)
}

export function getPosition(element) {
    const clientRect = element.getBoundingClientRect();
    return {
        left: clientRect.left + window.pageXOffset,
        top: clientRect.top + window.pageYOffset
    };
}