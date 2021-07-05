const patched = require('./patched')
const GO_WIRED_CONNECTED = "go-wired-connected";
const GO_WIRED_COMPONENT_ID = "go-wired-component-id";
const EVENT_WIRED_DOM_COMPONENT_ID_KEY = "cid";
const EVENT_WIRED_DOM_PATCHES_KEY = "i";
const EVENT_WIRED_DOM_TYPE_KEY = "t";
const EVENT_WIRED_DOM_CONTENT_KEY = "c";
const EVENT_WIRED_DOM_ATTR_KEY = "a";
const EVENT_WIRED_DOM_SELECTOR_KEY = "s";
const EVENT_WIRED_DOM_INDEX_KEY = "i";

const handlePatches= {
    "{{ .Enum.PatchedSetAttr }}": patched.setAttr,
    "{{ .Enum.PatchedRemoveAttr }}": patched.removeAttr,
    "{{ .Enum.PatchedReplace }}": patched.replace,
    "{{ .Enum.PatchedRemove }}": patched.remove,
    "{{ .Enum.PatchedAddHTML }}": patched.setInnerHTML,
    "{{ .Enum.PatchedAppend }}": patched.append,
    "{{ .Enum.PatchedMove }}": patched.move,
};

const goWired = {
    server: createConnection(),

    handlers: [],

    once: createOnceEmitter(),

    getWiredComponent(id) {
        return document.querySelector(
            ["*[", GO_WIRED_COMPONENT_ID, "=", id, "]"].join("")
        );
    },

    on(name, handler) {
        const newSize = this.handlers.push({
            name,
            handler,
        })
        return newSize - 1;
    },

    findHandler(name) {
        return this.handlers.filter((i) => i.name === name);
    },

    emit(name, message) {
        for (const handler of this.findHandler(name)) {
            handler.handler(message);
        }
    },

    off(index) {
        this.handlers.splice(index, 1);
    },

    send(message) {
        goWired.server.send(JSON.stringify(message));
    },

    connectChildren(viewElement) {
        const wiredChildren = viewElement.querySelectorAll(
            "*[" + GO_WIRED_COMPONENT_ID + "]"
        );

        wiredChildren.forEach((child) => {
            this.connectElement(child);
        });
    },

    connectElement(viewElement) {
        if (typeof viewElement === "string") {
            console.warn("is string")
            return;
        }

        if (!isElement(viewElement)) {
            console.warn("not element")
            return;
        }

        const connectedElements = []

        const clickElements = findWiredClicksFromElement(viewElement);
        clickElements.forEach(function (element) {

            const componentId = getComponentIdFromElement(element);

            element.addEventListener("click", function (_) {
                goWired.send({
                    name: "{{ .Enum.EventWiredMethod }}",
                    component_id: componentId,
                    method_name: element.getAttribute("go-wired-click"),
                    method_data: dataFromElementAttributes(element),
                });
            });

            connectedElements.push(element)
        });

        const keydownElements = findWiredKeyDownFromElement(viewElement);
        keydownElements.forEach(function (element) {

            const componentId = getComponentIdFromElement(element);
            const method = element.getAttribute("go-wired-keydown");

            const attrs = element.attributes;
            let filterKeys = [];
            for (let i = 0; i < attrs.length; i++) {
                if (
                    attrs[i].name === "go-wired-key" ||
                    attrs[i].name.startsWith("go-wired-key-")
                ) {
                    filterKeys.push(attrs[i].value);
                }
            }

            element.addEventListener("keydown", function (event) {
                const code = String(event.code);
                let hit = true;

                if (filterKeys.length !== 0) {
                    hit = false;
                    for (let i = 0; i < filterKeys.length; i++) {
                        if (filterKeys[i] === code) {
                            hit = true;

                            break;
                        }
                    }
                }

                if (hit) {
                        goWired.send({
                        name: "{{ .Enum.EventWiredMethod }}",
                        component_id: componentId,
                        method_name: method,
                        method_data: dataFromElementAttributes(element),
                        dom_event: {
                            keyCode: code,
                        },
                    });
                }
            });

            connectedElements.push(element)
        });

        const wiredInputs = findWiredInputsFromElement(viewElement);
        wiredInputs.forEach(function (element) {

            const type = element.getAttribute("type");
            const componentId = getComponentIdFromElement(element);

            element.addEventListener("input", function (_) {
                let value = element.value;

                if (type === "checkbox") {
                    value = element.checked;
                }

                goWired.send({
                    name: "{{ .Enum.EventWiredInput }}",
                    component_id: componentId,
                    key: element.getAttribute("go-wired-input"),
                    value: String(value),
                });
            });

            connectedElements.push(element)
        });


        for( const el of connectedElements ) {
            el.setAttribute(GO_WIRED_CONNECTED, true);
        }
    },

    connect(id) {
        const element = goWired.getWiredComponent(id);

        goWired.connectElement(element);

        goWired.on(
            "{{ .Enum.EventWiredDom }}",
            function handleWiredDom(message) {
                if (id === message[EVENT_LIVE_DOM_COMPONENT_ID_KEY]) {
                    for (const patched of message[
                        EVENT_LIVE_DOM_DIFFS_KEY
                        ]) {
                        const type = patched[EVENT_LIVE_DOM_TYPE_KEY];
                        const content = patched[EVENT_LIVE_DOM_CONTENT_KEY];
                        const attr = patched[EVENT_LIVE_DOM_ATTR_KEY];
                        const selector = patched[EVENT_LIVE_DOM_SELECTOR_KEY];
                        const index = patched[EVENT_LIVE_DOM_INDEX_KEY]

                        const element = document.querySelector(selector);

                        if (!element) {
                            console.error("Element not found", selector);
                            return;
                        }

                        handlePatches[type](
                            {
                                content: content,
                                attr: attr,
                                index: index
                            },
                            element,
                            id
                        );
                    }
                }
            }
        );
    },
};

goWired.once.on("WS_CONNECTION_OPEN", () => {
    goWired.on("{{ .Enum.EventWiredConnectElement }}", (message) => {
        const cid = message[EVENT_LIVE_DOM_COMPONENT_ID_KEY];
        goWired.connect(cid);
    });
    goWired.on("{{ .Enum.EventWiredError }}", (message) => {
        console.error("message", message.m)
        if (
            message.m ===
            '{{ index .EnumWiredError ` + "`WiredErrorSessionNotFound`" + `}}'
        ) {
            window.location.reload(false);
        }
    });
});

goWired.server.onmessage = (rawMessage) => {
    try {
        const message = JSON.parse(rawMessage.data);
        goWired.emit(message.t, message);
    } catch (e) {
        console.log("Error", e);
        console.log("Error message", rawMessage.data);
    }
};

goWired.server.onopen = () => {
    goWired.once.emit("WS_CONNECTION_OPEN");
};

function createConnection() {
    const path = [];

    if (window.location.protocol === "https:") {
        path.push("wss");
    } else {
        path.push("ws");
    }

    path.push("://", window.location.host, "/ws");

    return new WebSocket(path.join(""));
}

function createOnceEmitter() {
    const handlers = {};
    const createHandler = (name, called) => {
        handlers[name] = {
            called,
            cbs: [],
        };

        return handlers[name];
    };

    return {
        on(name, cb) {
            let handler = handlers[name];

            if (!handler) {
                handler = createHandler(name, false);
            }

            handler.cbs.push(cb);
        },
        emit(name, ...attrs) {
            const handler = handlers[name];

            if (!handler) {
                createHandler(name, true);
                return;
            }

            for (const cb of handler.cbs) {
                cb();
            }
        },
    };
}

const findWiredInputsFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-wired-input]:not([", GO_WIRED_CONNECTED, "])"].join("")
    );
};

const findWiredClicksFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-wired-click]:not([", GO_WIRED_CONNECTED, "])"].join("")
    );
};

const findWiredKeyDownFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-wired-keydown]:not([", GO_WIRED_CONNECTED, "])"].join("")
    );
};

const dataFromElementAttributes = (el) => {
    const attrs = el.attributes;
    let data = {};
    for (let i = 0; i < attrs.length; i++) {
        if (attrs[i].name.startsWith("go-wired-data-")) {
            data[attrs[i].name.substring(13)] = attrs[i].value;
        }
    }

    return data;
};

function getElementChild(element, index) {
    let el = element.firstElementChild;

    while (index > 0) {
        if (!el) {
            console.error("Element not found in path", element);
            return;
        }

        el = el.nextSibling;

        if (el.nodeType !== Node.ELEMENT_NODE) {
            continue
        }

        index--;
    }

    return el;
}

function isElement(o) {
    return typeof HTMLElement === "object"
        ? o instanceof HTMLElement //DOM2
        : o &&
        typeof o === "object" &&
        o.nodeType === 1 &&
        typeof o.nodeName === "string";
}

const getComponentIdFromElement = (element) => {
    const attr = element.getAttribute("go-wired-component-id");
    if (attr) {
        return attr;
    }

    if (element.parentElement) {
        return getComponentIdFromElement(element.parentElement);
    }

    return undefined;
};
