const dropZone = document.getElementById('drop-zone');
const dropZoneText = document.getElementById('drop-zone-text');
const uploadContainer = document.getElementById('upload-container');
let fileData;

['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    let a = (event) => {
        event.preventDefault()
    }

    dropZone.addEventListener(eventName, a)
});

dropZone.addEventListener('drop', handleDrop, false);

/**
 *
 * @param e {DragEvent}
 */
function handleDrop(e) {
    const files = e.dataTransfer.files;
    fileData = files[0]
    handleFiles(fileData);
}

/**
 *
 * @param file {File}
 */
function handleFiles(file) {
    if (file.type !== "application/pdf"){
        return
    }

    dropZoneText.textContent = `${file.name}`;
}

function toggleUploadPopup() {
    if (uploadContainer.style.display === "none") {
        uploadContainer.style.display = "flex"
        return
    }

    uploadContainer.style.display = "none"
    reset()
}

async function uploadContents() {
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
    fileData = null;
    dropZoneText.textContent = "Drag and drop pdf file here"
}

/**
 *
 * @param data {string}
 */
function sendData(data){

}