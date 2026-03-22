let documentModule = (function () {
        function deleteDocument(documentUUID) {
            return fetch(`/user/documents/${documentUUID}`, {
                credentials: "same-origin", method: "DELETE", headers: {
                    "Content-Type": "application/json",
                }
            });
        }

        function uploadDocument(documentBase64String, documentTitle, ownerType = "1") {
            return fetch("/user/upload", {
                credentials: "same-origin", method: "POST", headers: {
                    "Content-Type": "application/json",
                }, body: JSON.stringify({
                    documentBase64String: documentBase64String, documentTitle: documentTitle, ownerType: ownerType
                })
            });
        }

        /**
         * @param promise {Promise<Response>}
         * @param callback {function()}
         * @param errorMsg {String}
         */
        function completeFunctionThenCallback(promise, callback = refreshBasedOnHTTPCode, errorMsg = "FailedToResolvePromise") {
            if (!promise.then) {
                try {
                    callback(promise)
                } catch (e) {
                    notificationsModule.createError("Error", errorMsg);
                    console.error(e)
                }

                return;
            }

            promise.then(callback).catch((err) => {
                notificationsModule.createError("Error", errorMsg);
                console.error(err)
            })
        }

        /**
         *
         * @param res {ResponseInit}
         */
        function refreshBasedOnHTTPCode(res) {
            if (!res.status) {
                console.error("Unable to get status code form promise resolve.")
                return;
            }

            let value = res.status;
            if (value === 200) {
                return;
            }

            if (value === 302) {
                window.location.reload();
                return;
            }

            notificationsModule.createError("Error", "Something unexpected has happened!");
        }

        return {
            deleteDocument: deleteDocument,
            uploadDocument: uploadDocument,
            completeFunctionThenCallback: completeFunctionThenCallback
        }
    }

)()