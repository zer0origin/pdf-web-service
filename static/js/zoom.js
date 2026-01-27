var zoomModule = (function() {
    let currentZoomLevel = 1;
    let defaultWidth = null
    let defaultHeight = null
    let zoomCounterHtmlElement = document.getElementById("zoom-info");

    /**
     * Gets the array of images that the user can currently see.
     * @returns {HTMLCollectionOf<Element>}
     */
    function getImages(){
        return document.getElementsByClassName("viewerImage");
    }

    /**
     * Updates the default settings of the zoom module.
     * @param height {Number}
     * @param width {Number}
     * @param zoomCounter {HTMLElement}
     */
    function setDefaults(height = null, width = null, zoomCounter = null){
        if (height) {
            defaultHeight = height
        }

        if (width) {
            defaultWidth = width
        }

        if (zoomCounter){
            zoomCounterHtmlElement = zoomCounter
        }

        let images = getImages();
        for (let i = 0; i < images.length; i++) {
            applyImageSize(images[i], width, height)
        }
    }

    /**
     * Sets and applies the zoom level to images on the screen.
     * @param zoomLevel {number}
     */
    function setZoomLevel(zoomLevel) {
        if (zoomLevel > 3){
            zoomLevel = 3
        }

        if (zoomLevel < 0.1){
            zoomLevel = 0.1
        }

        let images = getImages();
        currentZoomLevel = Number(zoomLevel.toFixed(2))
        for (let i = 0; i < images.length; i++) {
            let image = images.item(i);
            applyImageSize(image, defaultWidth, defaultHeight)
        }

        zoomCounterHtmlElement.textContent = `${currentZoomLevel}`
        fireOnChange();
    }

    /**
     * Apply the current zoom level to a set of images, given a default height and width.
     * @param image {Element | null}
     * @param width {Number}
     * @param height {Number}
     */
    function applyImageSize(image, width, height){
        image.width = `${width * currentZoomLevel}`
        image.height = `${height * currentZoomLevel}`
    }

    /**
     * Get the current zoom level
     * @returns {number}
     */
    function getZoomLevel(){
        return currentZoomLevel
    }

    /**
     * Increase the zoom
     */
    function increaseZoom(){
        setZoomLevel(currentZoomLevel + 0.1)
    }

    /**
     * Decrease the zoom
     */
    function decreaseZoom(){
        setZoomLevel(currentZoomLevel - 0.1)
    }

    /**
     * Register a function to be executed, whenever there is a zoom change event.
     * @param func {() => void}
     */
    function registerZoomChange(func){
        onChangeListener.push(func)
    }

    const onChangeListener = []
    /**
     * Fire the zoom change handlers.
     */
    function fireOnChange(){
        for (let i = 0; i < onChangeListener.length; i++) {
            let func = onChangeListener[i];
            func();
        }
    }

    return {
        setDefaults: setDefaults,
        getZoomLevel: getZoomLevel,
        increaseZoom: increaseZoom,
        decreaseZoom: decreaseZoom,
        setZoomLevel: setZoomLevel,
        registerZoomChange: registerZoomChange,
        getImages: getImages,
        applyImageSize: applyImageSize
   }
})();