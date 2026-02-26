window.dropZone = document.getElementById('drop-zone');
window.dropZoneText = document.getElementById('drop-zone-text');
window.uploadContainer = document.getElementById('upload-container');
window.filesListElement = document.getElementById('fileListButton');

var uploadModule = (function () {
    function setup() {
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            let a = (event) => {
                event.preventDefault()
            }

            if (eventName === "drop") {
                window.dropZone.addEventListener(eventName, handleDrop, false);
            }

            window.dropZone.addEventListener(eventName, a)
        });

        reset()
    }

    async function onClick() {
        if (!isFileSelected()) {
            openFileBrowser();
            return
        }

        let fileData = window.filesListElement.files[0];
        if (!fileData) {
            reset()
            return;
        }

        if (fileData.type !== "application/pdf") {
            reset()
            return;
        }

        try {
            let documentBase64String = await toBase64(fileData)
            let documentTitle = fileData.name.slice(0, fileData.name.length - 4);
            documentModule.completePromiseThenCallback(
                documentModule.uploadDocument(documentBase64String, documentTitle),
                documentModule.completePromiseThenCallback
            )
        } catch (error) {
            console.error("Error preparing file for upload:", error);
        } finally {
            reset();
        }
    }

    /**
     *
     * @param e {DragEvent}
     */
    function handleDrop(e) {
        const files = e.dataTransfer.files;
        const file = files[0]

        if (file.type !== "application/pdf") {
            return
        }

        window.filesListElement.files = e.dataTransfer.files
        updateText(e)
    }

    function updateText(event) {
        console.log(event)
        const file = window.filesListElement.files[0]

        if (!file) {
            return
        }

        window.dropZoneText.textContent = `${file.name}`;
    }

    function hide() {
        window.uploadContainer.style.display = "none"
    }

    function show() {
        window.uploadContainer.style.display = "flex"
    }

    function toBase64(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result.toString().replace(/^data:(.*,)?/, ''));
            reader.onerror = reject;
        });
    }

    function reset() {
        window.filesListElement.value = ""
        window.dropZoneText.textContent = "Drag and drop pdf file here"
        hide()
    }

    function toggleUploadPopup() {
        if (window.uploadContainer.style.display === "none") {
            show()
            return
        }

        hide()
        reset()
    }

    function isFileSelected() {
        return window.filesListElement.files[0] !== undefined
    }

    function openFileBrowser() {
        window.filesListElement.click();
    }

    return {
        setup: setup,
        onClick: onClick,
        toggleUploadPopup: toggleUploadPopup,
        updateText: updateText
    }
})()
uploadModule.setup()