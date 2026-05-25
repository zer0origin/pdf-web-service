var ExtractModule = (function () {
    async function basic() {
        await apiModule.save();
        let selections = apiModule.convertSelectionMapToDTO(selectionsModule.map, true)
        let ids = selections.map(e => e.id)

        //create request!
        let req = {DocumentUid: getDocumentId(), Uids: ids}
        console.log(req)

        let res = await apiModule.sendBasicExtractRequest(req);
        console.log(res)
    }

    /**
     * @returns {String}
     */
    function getDocumentId() {
        return document.getElementById("viewer").attributes["documentId"].nodeValue;
    }

    return {
        basic: basic
    }
})();

/**
 * Worked on creating the extract controller logic in the web that will handle including the OwnerUID based on a token.
 * Need to work on the frontend js script that will handle getting all the selection IDs and sending them as a request to JESR,
 *
 *
 * Need to figure out what to do with the results.
 */