
var system = require("system")
var webpage = require('webpage');
var webserver = require('webserver');

/*
 * HTTP API
 */

// Serves RPC API.
var server = webserver.create();
// server.listen(system.env["PORT"], function(request, response) {
server.listen(system.args[1], function(request, response) {
	try {
		switch (request.url) {
			case '/ping': return handlePing(request, response);
			case '/webpage/CanGoBack': return handleWebpageCanGoBack(request, response);
			case '/webpage/CanGoForward': return handleWebpageCanGoForward(request, response);
			case '/webpage/ClipRect': return handleWebpageClipRect(request, response);
			case '/webpage/SetClipRect': return handleWebpageSetClipRect(request, response);
			case '/webpage/Cookies': return handleWebpageCookies(request, response);
			case '/webpage/SetCookies': return handleWebpageSetCookies(request, response);
			case '/webpage/CustomHeaders': return handleWebpageCustomHeaders(request, response);
			case '/webpage/SetCustomHeaders': return handleWebpageSetCustomHeaders(request, response);
			case '/webpage/Create': return handleWebpageCreate(request, response);
			case '/webpage/Content': return handleWebpageContent(request, response);
			case '/webpage/SetContent': return handleWebpageSetContent(request, response);
			case '/webpage/FocusedFrameName': return handleWebpageFocusedFrameName(request, response);
			case '/webpage/FrameContent': return handleWebpageFrameContent(request, response);
			case '/webpage/SetFrameContent': return handleWebpageSetFrameContent(request, response);
			case '/webpage/FrameName': return handleWebpageFrameName(request, response);
			case '/webpage/FramePlainText': return handleWebpageFramePlainText(request, response);
			case '/webpage/FrameTitle': return handleWebpageFrameTitle(request, response);
			case '/webpage/FrameURL': return handleWebpageFrameURL(request, response);
			case '/webpage/FrameCount': return handleWebpageFrameCount(request, response);
			case '/webpage/FrameNames': return handleWebpageFrameNames(request, response);
			case '/webpage/LibraryPath': return handleWebpageLibraryPath(request, response);
			case '/webpage/SetLibraryPath': return handleWebpageSetLibraryPath(request, response);
			case '/webpage/NavigationLocked': return handleWebpageNavigationLocked(request, response);
			case '/webpage/SetNavigationLocked': return handleWebpageSetNavigationLocked(request, response);
			case '/webpage/OfflineStoragePath': return handleWebpageOfflineStoragePath(request, response);
			case '/webpage/OfflineStorageQuota': return handleWebpageOfflineStorageQuota(request, response);
			case '/webpage/OwnsPages': return handleWebpageOwnsPages(request, response);
			case '/webpage/SetOwnsPages': return handleWebpageSetOwnsPages(request, response);
			case '/webpage/PageWindowNames': return handleWebpagePageWindowNames(request, response);
			case '/webpage/Pages': return handleWebpagePages(request, response);
			case '/webpage/PaperSize': return handleWebpagePaperSize(request, response);
			case '/webpage/SetPaperSize': return handleWebpageSetPaperSize(request, response);
			case '/webpage/PlainText': return handleWebpagePlainText(request, response);
			case '/webpage/ScrollPosition': return handleWebpageScrollPosition(request, response);
			case '/webpage/SetScrollPosition': return handleWebpageSetScrollPosition(request, response);
			case '/webpage/Settings': return handleWebpageSettings(request, response);
			case '/webpage/SetSettings': return handleWebpageSetSettings(request, response);
			case '/webpage/Title': return handleWebpageTitle(request, response);
			case '/webpage/URL': return handleWebpageURL(request, response);
			case '/webpage/ViewportSize': return handleWebpageViewportSize(request, response);
			case '/webpage/SetViewportSize': return handleWebpageSetViewportSize(request, response);
			case '/webpage/WindowName': return handleWebpageWindowName(request, response);
			case '/webpage/ZoomFactor': return handleWebpageZoomFactor(request, response);
			case '/webpage/SetZoomFactor': return handleWebpageSetZoomFactor(request, response);

			case '/webpage/AddCookie': return handleWebpageAddCookie(request, response);
			case '/webpage/ClearCookies': return handleWebpageClearCookies(request, response);
			case '/webpage/DeleteCookie': return handleWebpageDeleteCookie(request, response);
			case '/webpage/Open': return handleWebpageOpen(request, response);
			case '/webpage/Close': return handleWebpageClose(request, response);
			case '/webpage/EvaluateAsync': return handleWebpageEvaluateAsync(request, response);
			case '/webpage/EvaluateJavaScript': return handleWebpageEvaluateJavaScript(request, response);
			case '/webpage/Evaluate': return handleWebpageEvaluate(request, response);
			case '/webpage/Page': return handleWebpagePage(request, response);
			case '/webpage/GoBack': return handleWebpageGoBack(request, response);
			case '/webpage/GoForward': return handleWebpageGoForward(request, response);
			case '/webpage/Go': return handleWebpageGo(request, response);
			case '/webpage/IncludeJS': return handleWebpageIncludeJS(request, response);
			case '/webpage/InjectJS': return handleWebpageInjectJS(request, response);
			case '/webpage/Reload': return handleWebpageReload(request, response);
			case '/webpage/RenderBase64': return handleWebpageRenderBase64(request, response);
			case '/webpage/Render': return handleWebpageRender(request, response);
			case '/webpage/SendMouseEvent': return handleWebpageSendMouseEvent(request, response);
			case '/webpage/SendKeyboardEvent': return handleWebpageSendKeyboardEvent(request, response);
			case '/webpage/SetContentAndURL': return handleWebpageSetContentAndURL(request, response);
			case '/webpage/Stop': return handleWebpageStop(request, response);
			case '/webpage/SwitchToFocusedFrame': return handleWebpageSwitchToFocusedFrame(request, response);
			case '/webpage/SwitchToFrameName': return handleWebpageSwitchToFrameName(request, response);
			case '/webpage/SwitchToFramePosition': return handleWebpageSwitchToFramePosition(request, response);
			case '/webpage/SwitchToMainFrame': return handleWebpageSwitchToMainFrame(request, response);
			case '/webpage/SwitchToParentFrame': return handleWebpageSwitchToParentFrame(request, response);
			case '/webpage/UploadFile': return handleWebpageUploadFile(request, response);
			default: return handleNotFound(request, response);
		}
	} catch(e) {
		response.statusCode = 500;
		response.write(JSON.stringify({url: request.url, error: e.message}));
		response.closeGracefully();
	}
});

function handlePing(request, response) {
	response.statusCode = 200;
	response.write('ok');
	response.closeGracefully();
}

function handleWebpageCanGoBack(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.canGoBack}));
	response.closeGracefully();
}

function handleWebpageCanGoForward(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.canGoForward}));
	response.closeGracefully();
}

function handleWebpageClipRect(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.clipRect}));
	response.closeGracefully();
}

function handleWebpageSetClipRect(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.clipRect = msg.rect;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageCookies(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.cookies}));
	response.closeGracefully();
}

function handleWebpageSetCookies(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.cookies = msg.cookies;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageCustomHeaders(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.customHeaders}));
	response.closeGracefully();
}

function handleWebpageSetCustomHeaders(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.customHeaders = msg.headers;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageCreate(request, response) {
	var ref = createRef(webpage.create());
	response.statusCode = 200;
	response.write(JSON.stringify({ref: ref}));
	response.closeGracefully();
}

function handleWebpageOpen(request, response) {
	// response.write(JSON.stringify(request));
	// response.write("post:" + request.post);
  var msg = request.post;
  if(typeof request.post === 'string'){
    msg = JSON.parse(request.post);
  }
	var page = ref(msg.ref);

  page.onResourceRequested = function(requestData, pageRequest) {
    var resURL = requestData['url']
    if ((/http:\/\/.+?\.css(\?.+)?$/gi).test(resURL)) {
      pageRequest.abort();
    }else if ((/http:\/\/.+?\.png(\?.+)?$/gi).test(resURL)) {
      pageRequest.abort();
    }else if ((/http:\/\/.+?\.gif(\?.+)?$/gi).test(resURL)) {
      pageRequest.abort();
    }else if ((/http:\/\/.+?\.jpeg(\?.+)?$/gi).test(resURL)) {
      pageRequest.abort();
    }else if ((/http:\/\/.+?\.jpg(\?.+)?$/gi).test(resURL)) {
      pageRequest.abort();
    }
  };

	// response.write("page open");
	page.open(msg.url, function(status) {
		response.write(JSON.stringify({status: status}));
		response.closeGracefully();
	})
}

function handleWebpageContent(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.content}));
	response.closeGracefully();
}

function handleWebpageSetContent(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.content = msg.content;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageFocusedFrameName(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.focusedFrameName}));
	response.closeGracefully();
}

function handleWebpageFrameContent(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.frameContent}));
	response.closeGracefully();
}

function handleWebpageSetFrameContent(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.frameContent = msg.content;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageFrameName(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.frameName}));
	response.closeGracefully();
}

function handleWebpageFramePlainText(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.framePlainText}));
	response.closeGracefully();
}

function handleWebpageFrameTitle(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.frameTitle}));
	response.closeGracefully();
}

function handleWebpageFrameURL(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.frameUrl}));
	response.closeGracefully();
}

function handleWebpageFrameCount(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.framesCount}));
	response.closeGracefully();
}

function handleWebpageFrameNames(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.framesName}));
	response.closeGracefully();
}

function handleWebpageLibraryPath(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.libraryPath}));
	response.closeGracefully();
}

function handleWebpageSetLibraryPath(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.libraryPath = msg.path;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageNavigationLocked(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.navigationLocked}));
	response.closeGracefully();
}

function handleWebpageSetNavigationLocked(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.navigationLocked = msg.value;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageOfflineStoragePath(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.offlineStoragePath}));
	response.closeGracefully();
}

function handleWebpageOfflineStorageQuota(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.offlineStorageQuota}));
	response.closeGracefully();
}

function handleWebpageOwnsPages(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.ownsPages}));
	response.closeGracefully();
}

function handleWebpageSetOwnsPages(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.ownsPages = msg.value;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpagePageWindowNames(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.pagesWindowName}));
	response.closeGracefully();
}

function handleWebpagePages(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	var refs = page.pages.map(function(p) { return createRef(p); })
	response.write(JSON.stringify({refs: refs}));
	response.closeGracefully();
}

function handleWebpagePaperSize(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.paperSize}));
	response.closeGracefully();
}

function handleWebpageSetPaperSize(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.paperSize = msg.size;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpagePlainText(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.plainText}));
	response.closeGracefully();
}

function handleWebpageScrollPosition(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	var pos = page.scrollPosition;
	response.write(JSON.stringify({top: pos.top, left: pos.left}));
	response.closeGracefully();
}

function handleWebpageSetScrollPosition(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.scrollPosition = {top: msg.top, left: msg.left};
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSettings(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({settings: page.settings}));
	response.closeGracefully();
}

function handleWebpageSetSettings(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.settings = msg.settings;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageTitle(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.title}));
	response.closeGracefully();
}

function handleWebpageURL(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.url}));
	response.closeGracefully();
}

function handleWebpageViewportSize(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	var viewport = page.viewportSize;
	response.write(JSON.stringify({width: viewport.width, height: viewport.height}));
	response.closeGracefully();
}

function handleWebpageSetViewportSize(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.viewportSize = {width: msg.width, height: msg.height};
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageWindowName(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.windowName}));
	response.closeGracefully();
}

function handleWebpageZoomFactor(request, response) {
	var page = ref(JSON.parse(request.post).ref);
	response.write(JSON.stringify({value: page.zoomFactor}));
	response.closeGracefully();
}

function handleWebpageSetZoomFactor(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.zoomFactor = msg.value;
	response.write(JSON.stringify({}));
	response.closeGracefully();
}


function handleWebpageAddCookie(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.addCookie(msg.cookie);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpageClearCookies(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.clearCookies();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageDeleteCookie(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.deleteCookie(msg.name);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpageClose(request, response) {
	var msg = JSON.parse(request.post);

	// Close page.
	var page = ref(msg.ref);
	page.close();
	delete(refs, msg.ref);

	// Close and dereference owned pages.
	for (var i = 0; i < page.pages.length; i++) {
		page.pages[i].close();
		deleteRef(page.pages[i]);
	}

	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageEvaluateAsync(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.evaluateAsync(msg.script, msg.delay);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageEvaluateJavaScript(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.evaluateJavaScript(msg.script);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpageEvaluate(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.evaluate(msg.script);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpagePage(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var p = page.getPage(msg.name);

	if (p === null) {
		response.write(JSON.stringify({}));
	} else {
		response.write(JSON.stringify({ref: createRef(p)}));
	}
	response.closeGracefully();
}

function handleWebpageGoBack(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.goBack();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageGoForward(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.goForward();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageGo(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.go(msg.index);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageIncludeJS(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.includeJs(msg.url, function() {
		response.write(JSON.stringify({}));
		response.closeGracefully();
	});
}

function handleWebpageInjectJS(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.injectJs(msg.filename);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpageReload(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.reload();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageRenderBase64(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	var returnValue = page.renderBase64(msg.format);
	response.write(JSON.stringify({returnValue: returnValue}));
	response.closeGracefully();
}

function handleWebpageRender(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.render(msg.filename, {format: msg.format, quality: msg.quality});
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSendMouseEvent(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.sendEvent(msg.eventType, msg.mouseX, msg.mouseY, msg.button);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSendKeyboardEvent(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.sendEvent(msg.eventType, msg.key, null, null, msg.modifier);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSetContentAndURL(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.setContent(msg.content, msg.url);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageStop(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.stop();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSwitchToFocusedFrame(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.switchToFocusedFrame();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSwitchToFrameName(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.switchToFrame(msg.name);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSwitchToFramePosition(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.switchToFrame(msg.position);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSwitchToMainFrame(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.switchToMainFrame();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageSwitchToParentFrame(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.switchToParentFrame();
	response.write(JSON.stringify({}));
	response.closeGracefully();
}

function handleWebpageUploadFile(request, response) {
	var msg = JSON.parse(request.post);
	var page = ref(msg.ref);
	page.uploadFile(msg.selector, msg.filename);
	response.write(JSON.stringify({}));
	response.closeGracefully();
}


function handleNotFound(request, response) {
	response.statusCode = 404;
	// response.write(JSON.stringify({error:"not found"}));
	response.write(JSON.stringify(request));
	response.closeGracefully();
}


/*
 * REFS
 */

// Holds references to remote objects.
var refID = 0;
var refs = {};

// Adds an object to the reference map and a ref object.
function createRef(value) {
	// Return existing reference, if one exists.
	for (var key in refs) {
		if (refs.hasOwnProperty(key)) {
			if (refs[key] === value) {
				return key
			}
		}
	}

	// Generate a new id for new references.
	refID++;
	refs[refID.toString()] = value;
	return {id: refID.toString()};
}

// Removes a reference to a value, if any.
function deleteRef(value) {
	for (var key in refs) {
		if (refs.hasOwnProperty(key)) {
			if (refs[key] === value) {
				delete(refs, key);
			}
		}
	}
}

// Returns a reference object by ID.
function ref(id) {
	return refs[id];
}
