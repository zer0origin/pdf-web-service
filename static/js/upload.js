window.dropZone = document.getElementById('drop-zone');
window.dropZoneText = document.getElementById('drop-zone-text');
window.uploadContainer = document.getElementById('upload-container');
window.filesListElement = document.getElementById('fileListButton');

setup()

function setup() {
    document.getElementById('customUploadButton').addEventListener('click', function () {
        if (!window.filesListElement.files[0]) {
            window.filesListElement.click();
        }
    });

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

function isFileSelected() {
    return window.filesListElement.files[0] !== undefined
}

function openFileBrowser() {
    window.filesListElement.click();
}

//USED BY HTMX
function shouldSendRequest(event) {
    if (!isFileSelected()) {
        openFileBrowser();
        event.preventDefault()
    }
}

//USED BY HTMX
function htmxUploadContents(event) {
    let fileData = window.filesListElement.files[0];
    if (!fileData) {
        return;
    }

    if (fileData.type !== "application/pdf"){ //TODO: WARNING EITHER CLIENT SIZE OR JS SIDE.
        reset()
    }

    event.detail.formData.append("documentBase64String", window.tempData);
    event.detail.formData.append("documentTitle", fileData.name.slice(0,fileData.name.length-4));
    event.detail.formData.append("ownerType", "1");
}

//USED BY HTMX
async function htmxConfirmEvent(event) {
    console.log(event);

    let fileData = window.filesListElement.files[0];
    if (!fileData) {
        return;
    }

    if (fileData.type !== "application/pdf"){
        reset()
    }

    // Prevent the request from being issued immediately
    event.preventDefault();

    try {
        window.tempData = await toBase64(fileData)
        event.detail.issueRequest();
    } catch (error) {
        console.error("Error preparing file for upload:", error);
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
    updateText()
}

function updateText() {
    const file = window.filesListElement.files[0]

    if (!file) {
        return
    }

    window.dropZoneText.textContent = `${file.name}`;
}

function toggleUploadPopup() {
    if (window.uploadContainer.style.display === "none") {
        show()
        return
    }

    hide()
    reset()
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