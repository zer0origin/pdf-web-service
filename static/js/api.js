class SelectionDTO {
    documentUUID
    coordinates
    pageKey

    constructor(uuid, bounds, key) {
        this.documentUUID = uuid
        this.coordinates = bounds
        this.pageKey = key
    }
}

var apiModule = (function (){
    /**
     *
     * @param data {Map<string, Array<Rectangle>>}
     */
    function convertSelectionMapToDTO(data ){
        mapDTO = []
        data.forEach((pageRectangleArray, key) => {
            pageRectangleArray.forEach((rectangle) => {
                p1 = rectangle.p1
                p2 = rectangle.p2

                if (!p1 || !p2){
                    return
                }

                let name = String(rectangle.imageDiv.id);
                let key = name.split("-")[1];

                let selectionDTO = new SelectionDTO(getDocumentId(), {X1: p1.x, Y1: p1.y, X2: p2.x, Y2: p2.y}, key)
                mapDTO.push(selectionDTO)
            })
        })

        return mapDTO
    }

    /**
     *
     * @param data {Map<string, Array<Rectangle>>}
     */
    function saveSelectionsToDatabase(data){
        let url = "/selection/bulk/"
        let promise = fetch(url, {
            method: "POST",
            cache: "no-cache",
            body: JSON.stringify(data)
        })

        promise.catch(reason => console.error(reason))
    }

    function loadSelectionsFromDatabase(){
        let url = `/selection?documentUUID=${getDocumentId()}`
        let promise = fetch(url, {
            method: "GET",
            cache: "default",
        })

        promise.then(res => res.json()).then(data => console.log(JSON.parse(data))).catch(reason => console.error(reason))
    }

    /**
     * @returns {string}
     */
    function getDocumentId(){
        return document.getElementById("viewer").attributes["documentId"].nodeValue;
    }

    return {
        getDocumentId: getDocumentId,
        convertSelectionMapToDTO: convertSelectionMapToDTO,
        saveSelectionsToDatabase: saveSelectionsToDatabase,
        load: loadSelectionsFromDatabase
    }
})()
