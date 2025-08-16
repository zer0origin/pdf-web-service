const dropZone = document.getElementById('drop-zone');
const dropZoneText = document.getElementById('drop-zone-text');
const uploadContainer = document.getElementById('upload-container');
/** @type {HTMLInputElement} */ const filesListElement = document.getElementById('fileListButton');

setup()

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

function hide(){
    uploadContainer.style.display = "none"
}

function show(){
    uploadContainer.style.display = "flex"
}

async function uploadContents() {
    let fileData = filesListElement.files[0]
    if (!fileData) {
        return
    }

    let dataToSend = await toBase64(fileData);
    console.log(dataToSend)
    sendData(dataToSend)
    reset()
}

/**
 *
 * @param file {File}
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

/**
 *
 * @param data {string}
 */
function sendData(data) {
    console.log("Sending data to the server")

    fetch("/user/upload", {
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        },
        method: "POST",
        body: JSON.stringify({
            documentBase64String: data,
            ownerType: 1,
        })
    }).then(r => {
        if (r.status !== 201){
            console.error("Server responded with an unexpected status code.")
        }

    });
}