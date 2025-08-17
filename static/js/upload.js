const dropZone = document.getElementById('drop-zone');
const dropZoneText = document.getElementById('drop-zone-text');
const uploadContainer = document.getElementById('upload-container');
/** @type {HTMLInputElement} */ const filesListElement = document.getElementById('fileListButton');
let tempData = ""

// setup()

function isFileSelected() {
    return filesListElement.files[0] !== undefined
}

function openFileBrowser() {
    filesListElement.click();
}

function shouldSendRequest(event) {
    if (!isFileSelected()) {
        openFileBrowser();
        event.preventDefault()
    }
}

function htmxUploadContents(event) {
    let fileData = filesListElement.files[0];
    if (!fileData) {
        return;
    }

    if (fileData.type !== "application/pdf"){
        reset()
    }

    event.detail.formData.append("documentBase64String", tempData);
    event.detail.formData.append("documentTitle", fileData.name);
    event.detail.formData.append("ownerType", "1");
}

async function htmxConfirmEvent(event) {
    console.log(event);

    let fileData = filesListElement.files[0];
    if (!fileData) {
        return;
    }

    if (fileData.type !== "application/pdf"){
        reset()
    }

    // Prevent the request from being issued immediately
    event.preventDefault();

    try {
        tempData = await toBase64(fileData)
        event.detail.issueRequest();
    } catch (error) {
        console.error("Error preparing file for upload:", error);
    }
}

function setup() {
    filesListElement.addEventListener("input", () => {
        console.log("INPUT UPDATED")
        updateText()
    })

    document.getElementById('customUploadButton').addEventListener('click', function () {
        if (!filesListElement.files[0]) {
            filesListElement.click();
        }
    });

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        let a = (event) => {
            event.preventDefault()
        }

        if (eventName === "drop") {
            dropZone.addEventListener(eventName, handleDrop, false);
        }

        dropZone.addEventListener(eventName, a)
    });

    reset()
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

    filesListElement.files = e.dataTransfer.files
    updateText()
}

function updateText() {
    const file = filesListElement.files[0]

    if (!file) {
        return
    }

    dropZoneText.textContent = `${file.name}`;
}

function toggleUploadPopup() {
    if (uploadContainer.style.display === "none") {
        show()
        return
    }

    hide()
    reset()
}

function hide() {
    uploadContainer.style.display = "none"
}

function show() {
    uploadContainer.style.display = "flex"
}

/**
 *
 * @param file {File}
 * @returns Promise<string>
 */
function toBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = () => resolve(reader.result.toString().replace(/^data:(.*,)?/, ''));
        reader.onerror = reject;
    });
}

function reset() {
    filesListElement.value = ""
    dropZoneText.textContent = "Drag and drop pdf file here"
    hide()
}