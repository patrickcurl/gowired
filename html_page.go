package golive

// Code automatically generated. DO NOT EDIT.
// > go run ci/create_html_page.go
var BasePageString = `<!DOCTYPE html>
<html lang="{{ .Lang }}">
  <head>
    <meta charset="UTF-8" />
    <title>{{ .Title }}</title>
    {{ .Head }}
  </head>

  <body>
    {{ .Body }}
  </body>

  <script type="application/javascript">
    const GO_LIVE_CONNECTED="go-live-connected",GO_LIVE_COMPONENT_ID="go-live-component-id",EVENT_LIVE_DOM_COMPONENT_ID_KEY="cid",EVENT_LIVE_DOM_INSTRUCTIONS_KEY="i",EVENT_LIVE_DOM_TYPE_KEY="t",EVENT_LIVE_DOM_CONTENT_KEY="c",EVENT_LIVE_DOM_ATTR_KEY="a",EVENT_LIVE_DOM_SELECTOR_KEY="s",EVENT_LIVE_DOM_INDEX_KEY="i",handleChange={"{{ .Enum.DiffSetAttr }}":handleDiffSetAttr,"{{ .Enum.DiffRemoveAttr }}":handleDiffRemoveAttr,"{{ .Enum.DiffReplace }}":handleDiffReplace,"{{ .Enum.DiffRemove }}":handleDiffRemove,"{{ .Enum.DiffSetInnerHTML }}":handleDiffSetInnerHTML,"{{ .Enum.DiffAppend }}":handleDiffAppend,"{{ .Enum.DiffMove }}":handleDiffMove},goLive={server:createConnection(),handlers:[],once:createOnceEmitter(),getLiveComponent(a){return document.querySelector(["*[",GO_LIVE_COMPONENT_ID,"=",a,"]"].join(""))},on(a,b){const c=this.handlers.push({name:a,handler:b});return c-1},findHandler(a){return this.handlers.filter(b=>b.name===a)},emit(a,b){for(const c of this.findHandler(a))c.handler(b)},off(a){this.handlers.splice(a,1)},send(a){goLive.server.send(JSON.stringify(a))},connectChildren(a){const b=a.querySelectorAll("*["+GO_LIVE_COMPONENT_ID+"]");b.forEach(a=>{this.connectElement(a)})},connectElement(a){if(typeof a=="string"){console.warn("is string");return}if(!isElement(a)){console.warn("not element");return}const b=[],c=findLiveClicksFromElement(a);c.forEach(function(a){const c=getComponentIdFromElement(a);a.addEventListener("click",function(b){goLive.send({name:"{{ .Enum.EventLiveMethod }}",component_id:c,method_name:a.getAttribute("go-live-click"),method_data:dataFromElementAttributes(a)})}),b.push(a)});const d=findLiveKeyDownFromElement(a);d.forEach(function(a){const e=getComponentIdFromElement(a),f=a.getAttribute("go-live-keydown"),c=a.attributes;let d=[];for(let a=0;a<c.length;a++)(c[a].name==="go-live-key"||c[a].name.startsWith("go-live-key-"))&&d.push(c[a].value);a.addEventListener("keydown",function(g){const c=String(g.code);let b=!0;if(d.length!==0){b=!1;for(let a=0;a<d.length;a++)if(d[a]===c){b=!0;break}}b&&goLive.send({name:"{{ .Enum.EventLiveMethod }}",component_id:e,method_name:f,method_data:dataFromElementAttributes(a),dom_event:{keyCode:c}})}),b.push(a)});const e=findLiveInputsFromElement(a);e.forEach(function(a){const c=a.getAttribute("type"),d=getComponentIdFromElement(a);a.addEventListener("input",function(e){let b=a.value;c==="checkbox"&&(b=a.checked),goLive.send({name:"{{ .Enum.EventLiveInput }}",component_id:d,key:a.getAttribute("go-live-input"),value:String(b)})}),b.push(a)});for(const a of b)a.setAttribute(GO_LIVE_CONNECTED,!0)},connect(a){const b=goLive.getLiveComponent(a);goLive.connectElement(b),goLive.on("{{ .Enum.EventLiveDom }}",function(b){if(a===b[EVENT_LIVE_DOM_COMPONENT_ID_KEY])for(const c of b[EVENT_LIVE_DOM_INSTRUCTIONS_KEY]){const f=c[EVENT_LIVE_DOM_TYPE_KEY],g=c[EVENT_LIVE_DOM_CONTENT_KEY],h=c[EVENT_LIVE_DOM_ATTR_KEY],d=c[EVENT_LIVE_DOM_SELECTOR_KEY],i=c[EVENT_LIVE_DOM_INDEX_KEY],e=document.querySelector(d);if(!e){console.error("Element not found",d);return}handleChange[f]({content:g,attr:h,index:i},e,a)}})}};goLive.once.on("WS_CONNECTION_OPEN",()=>{goLive.on("{{ .Enum.EventLiveConnectElement }}",a=>{const b=a[EVENT_LIVE_DOM_COMPONENT_ID_KEY];goLive.connect(b)}),goLive.on("{{ .Enum.EventLiveError }}",a=>{console.error("message",a.m),a.m==='{{ index .EnumLiveError ` + "`LiveErrorSessionNotFound`" + `}}'&&window.location.reload(!1)})}),goLive.server.onmessage=a=>{try{const b=JSON.parse(a.data);goLive.emit(b.t,b)}catch(b){console.log("Error",b),console.log("Error message",a.data)}},goLive.server.onopen=()=>{goLive.once.emit("WS_CONNECTION_OPEN")};function createConnection(){const a=[];return window.location.protocol==="https:"?a.push("wss"):a.push("ws"),a.push("://",window.location.host,"/ws"),new WebSocket(a.join(""))}function createOnceEmitter(){const a={},b=(b,c)=>(a[b]={called:c,cbs:[]},a[b]);return{on(d,e){let c=a[d];c||(c=b(d,!1)),c.cbs.push(e)},emit(c,...e){const d=a[c];if(!d){b(c,!0);return}for(const a of d.cbs)a()}}}const findLiveInputsFromElement=a=>a.querySelectorAll(["*[go-live-input]:not([",GO_LIVE_CONNECTED,"])"].join("")),findLiveClicksFromElement=a=>a.querySelectorAll(["*[go-live-click]:not([",GO_LIVE_CONNECTED,"])"].join("")),findLiveKeyDownFromElement=a=>a.querySelectorAll(["*[go-live-keydown]:not([",GO_LIVE_CONNECTED,"])"].join("")),dataFromElementAttributes=c=>{const a=c.attributes;let b={};for(let c=0;c<a.length;c++)a[c].name.startsWith("go-live-data-")&&(b[a[c].name.substring(13)]=a[c].value);return b};function getElementChild(b,c){let a=b.firstElementChild;while(c>0){if(!a){console.error("Element not found in path",b);return}if(a=a.nextSibling,a.nodeType!==Node.ELEMENT_NODE)continue;c--}return a}function isElement(a){return typeof HTMLElement=="object"?a instanceof HTMLElement:a&&typeof a=="object"&&a.nodeType===1&&typeof a.nodeName=="string"}function handleDiffSetAttr(c,b){const{attr:a}=c;a.Name==="value"&&b.value?b.value=a.Value:b.setAttribute(a.Name,a.Value)}function handleDiffRemoveAttr(a,b){const{attr:c}=a;b.removeAttribute(c.Name)}function handleDiffReplace(d,a){const{content:e}=d,b=document.createElement("div");b.innerHTML=e;const c=a.parentElement;c.replaceChild(b.firstChild,a),goLive.connectElement(c)}function handleDiffRemove(c,a){const b=a.parentElement;b.removeChild(a)}function handleDiffSetInnerHTML(c,a){let{content:b}=c;if(b===void 0&&(b=""),a.nodeType===Node.TEXT_NODE){a.textContent=b;return}a.innerHTML=b,goLive.connectElement(a)}function handleDiffAppend(c,a){const{content:d}=c,b=document.createElement("div");b.innerHTML=d;const e=b.firstChild;a.appendChild(e),goLive.connectElement(a)}function handleDiffMove(c,a){const b=a.parentNode;b.removeChild(a);const d=getElementChild(b,c.index);b.replaceChild(a,d)}const getComponentIdFromElement=a=>{const b=a.getAttribute("go-live-component-id");return b?b:a.parentElement?getComponentIdFromElement(a.parentElement):void 0}
  </script>
</html>
`