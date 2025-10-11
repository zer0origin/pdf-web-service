var notificationsModule = (function () {
    function listNotification(header = 'Success', message = 'Your changes have been saved'){
        let template = document.getElementById("notification-template")
        let node = document.importNode(template.content.querySelector("div"), true)

        const notif = node;
        const headerEl = node.querySelector('#notif-header');
        const textEl = node.querySelector('#notif-text');
        headerEl.textContent = header;
        textEl.textContent = message;
        notif.onanimationend = (e) => notif.remove();

        const exitBtn = notif.querySelector("#exit-button")
        exitBtn.onclick = (e) => notif.remove();

        notif.style.animation = 'none';
        void notif.offsetWidth;
        notif.style.animation = null;
        document.getElementById("notificationList").appendChild(node)
        notif.hidden = false;
    }

    return {
        create: listNotification,
    }
})()