var zoomModule = (function() {
    let currentZoomLevel = 1;
    const images = document.getElementsByClassName("viewerImage");
    let defaultWidth = Number(images.item(0).width);
    let defaultHeight = Number(images.item(0).height);
    let zoomCounter = document.getElementById("zoom-info");

    /**
     *
     * @type {[() => void]}
     */
    const onChangeListener = []

    /**
     *
     * @param zoomLevel {number}
     */
    function setZoomLevel(zoomLevel) {
        if (zoomLevel > 3){
            zoomLevel = 3
        }

        if (zoomLevel < 0.1){
            zoomLevel = 0.1
        }

        currentZoomLevel = Number(zoomLevel.toFixed(2))
        for (let i = 0; i < images.length; i++) {
            let image = images.item(i);
            image.width = `${defaultWidth * currentZoomLevel}`
            image.height = `${defaultHeight * currentZoomLevel}`
        }

        zoomCounter.textContent = `${currentZoomLevel}`
        fireOnChange();
    }

    function getZoomLevel(){
        return currentZoomLevel
    }

    function increaseZoom(){
        setZoomLevel(currentZoomLevel + 0.1)
    }

    function decreaseZoom(){
        setZoomLevel(currentZoomLevel - 0.1)
    }

    /**
     *
     * @param func {() => void}
     */
    function registerZoomChange(func){
        onChangeListener.push(func)
    }

    function fireOnChange(){
        for (let i = 0; i < onChangeListener.length; i++) {
            let func = onChangeListener[i];
            func();
        }
    }

    return {
        setZoomLevel: setZoomLevel,
        getZoomLevel: getZoomLevel,
        decreaseZoom: decreaseZoom,
        increaseZoom: increaseZoom,
        registerZoomChange: registerZoomChange,
    };
})();