window.dropZone = document.getElementById('drop-zone');
window.dropZoneText = document.getElementById('drop-zone-text');
window.uploadContainer = document.getElementById('upload-container');
window.filesListElement = document.getElementById('fileListButton');

setup()

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

    if (fileData.type !== "application/pdf") {
        reset()
    }

    event.detail.formData.append("documentBase64String", window.tempData);
    event.detail.formData.append("documentTitle", fileData.name.slice(0, fileData.name.length - 4));
    event.detail.formData.append("ownerType", "1");
}

//USED BY HTMX
async function htmxConfirmEvent(event) {
    let fileData = window.filesListElement.files[0];
    if (!fileData) {
        return;
    }

    if (fileData.type !== "application/pdf") {
        reset()
    }

    // Prevent the request from being issued immediately
    event.preventDefault();

    try {
        window.tempData = await toBase64(fileData)
        event.detail.issueRequest();
        reset();
    } catch (error) {
        console.error("Error preparing file for upload:", error);
    }
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
        let ownerType = "1";

        fetch("/user/upload", {
            credentials: "same-origin",
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                documentBase64String: documentBase64String,
                documentTitle: documentTitle,
                ownerType: ownerType
            })
        }).catch((err) => {
            notificationsModule.createError("Error", "Failed to upload document!");
            console.error(err)
        }).then((res) => res.status).then(value => {
            if (value === 200) {
                return;
            }

            if (value === 302) {
                window.location.reload();
                return;
            }

            notificationsModule.createError("Error", "Something unexpected has happened!");
        })


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