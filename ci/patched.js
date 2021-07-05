export modules.default = {
    append (message, el) {
        const { content } = message;

        const wrapper = document.createElement("div");
        wrapper.innerHTML = content;

        const child = wrapper.firstChild;
        el.appendChild(child);
        goWired.connectElement(el);
    },
    move (message, el) {
        const parent = el.parentNode
        parent.removeChild(el)

        const child = getElementChild(parent, message.index)
        parent.replaceChild(el, child)
    },
    removeAttr (message, el) {
        const { attr } = message;

        el.removeAttribute(attr.Name);
    },
    replace (message, el) {
        const { content } = message;

        const wrapper = document.createElement("div");
        wrapper.innerHTML = content;

        const parent = el.parentElement
        parent.replaceChild(wrapper.firstChild, el);

        goWired.connectElement(parent)
    },
    remove(message, el) {
        const parent = el.parentElement
        parent.removeChild(el);
    },
    setAttr (message, el) {
        const { attr } = message;

        if (attr.Name === "value" && el.value) {
            el.value = attr.Value;
        } else {
            el.setAttribute(attr.Name, attr.Value);
        }
    },
    setInnerHTML (message, el) {
         let { content } = message;

        if (content === undefined) {
            content = "";
        }

        if (el.nodeType === Node.TEXT_NODE) {
            el.textContent = content;
            return;
        }

        el.innerHTML = content;

        goWired.connectElement(el);
    }
}

