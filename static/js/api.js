SelectionDTO = class SelectionDTO {
    documentUUID
    /**
     * @type {{X1, X2, Y1, Y2}}
     */
    coordinates
    pageKey

    constructor(uuid, bounds, key) {
        this.documentUUID = uuid
        this.coordinates = bounds
        this.pageKey = key
    }
}

var apiModule = (function () {
    /**
     * @param data {Map<string, Array<Rectangle>>}
     * @param includeExternal Should selections from an external source be included in the DTO array?
     * @returns {SelectionDTO[]}
     */
    function convertSelectionMapToDTO(data, includeExternal = false) {
        let mapDTO = []
        data.forEach((pageRectangleArray, key) => {
            pageRectangleArray.filter(value => value.isExternal === includeExternal).forEach((rectangle) => {
                p1 = rectangle.p1
                p2 = rectangle.p2

                if (!p1 || !p2) {
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
     * @param data {*[]}
     */
    function saveSelectionsToDatabase(data) {
        let url = "/selection/bulk/"
        return fetch(url, {
            method: "POST", cache: "no-cache", body: JSON.stringify(data)
        });
    }

    /**
     * Load external selections from the database, and return a promise.
     * @returns {Promise<any>}
     */
    function loadSelectionsFromDatabase() {
        return new Promise((resolve, reject) => {
            let url = `/selection?documentUUID=${getDocumentId()}`
            let promise = fetch(url, {
                method: "GET", cache: "default",
            })

            promise.then(res => res.json()).then(data => {
                let selectionData = JSON.parse(data);
                selectionData.selections.map(t => selectionsModule.load(t.pageKey, new Point(t.coordinates.x1, t.coordinates.y1), new Point(t.coordinates.x2, t.coordinates.y2), t.selectionUUID))
                resolve(selectionData)
            }).catch(reason => reject(reason))
        })
    }

    let cooldown = false;
    async function save() {
        const data = selectionsModule.map;
        const dataToTransfer = convertSelectionMapToDTO(data);
        if (dataToTransfer.length <= 0) {
            console.error("Nothing new to save!")
            return
        }

        if (!cooldown) {
            let res = await saveSelectionsToDatabase(dataToTransfer).catch(reason => {
                notificationsModule.createError("Error", `Changes could not be saved!`)
                console.error(reason);
                return undefined;
            });

            if (!res){
                return
            }

            let body = await res.json();
            let respData = JSON.parse(body)

            dataToTransfer.forEach((sel, index) => {
                let transferred = data.get(sel.pageKey).find(rec => rec.p1.x === sel.coordinates.X1
                    && rec.p1.y === sel.coordinates.Y1
                    && rec.p2.x === sel.coordinates.X2
                    && rec.p2.y === sel.coordinates.Y2
                    && !rec.isExternal);

                transferred.isExternal = true
                transferred.id = respData.uids[index]
            })

            notificationsModule.create("Saved", "Your changes have been saved");
            cooldown = true;
            setTimeout(() => {
                cooldown = false;
            }, 1000)
        }
    }

    /**
     * @returns {string}
     */
    function getDocumentId() {
        return document.getElementById("viewer").attributes["documentId"].nodeValue;
    }

    return {
        getDocumentId: getDocumentId,
        convertSelectionMapToDTO: convertSelectionMapToDTO,
        saveSelectionsToDatabase: saveSelectionsToDatabase,
        load: loadSelectionsFromDatabase,
        save: save
    }
})
()
