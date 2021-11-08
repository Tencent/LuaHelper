const vscode = require('vscode');
const path = require('path');
const fs = require('fs');

var currentPanel;

function compileOther(srcPath, dstPath, name) {
    fs.copyFileSync(path.join(srcPath, name), path.join(dstPath, name));
}

function compileCss(srcPath, dstPath, name) {
    let css = fs.readFileSync(path.join(srcPath, name), 'utf8');
    css = css.split('\n').map(function (line) {
        if (line.match('color')) {
            return '';
        }
        return line;
    }).join('\n');
    fs.writeFileSync(path.join(dstPath, name), css);
}

function compileHtml(srcPath, dstPath, language, version, name) {
    let html = fs.readFileSync(path.join(srcPath, name), 'utf8');
    html = html.replace(/(<\/body>)/i, `
<SCRIPT>
    const vscode = acquireVsCodeApi();
    function saveState() {
        vscode.setState({
            language: "${language}",
            version: "${version}",
            file: "${name}",
            scrollTop: document.body.scrollTop + document.documentElement.scrollTop,
            scrollLeft: document.body.scrollLeft + document.documentElement.scrollLeft,
        });
    }
    function gotoAnchor(anchor) {
        for (const e of document.getElementsByName(anchor)) {
            e.scrollIntoView();
            saveState();
            break;
        }
    }
    for (const link of document.querySelectorAll('a[href^="#"]')) {
        link.addEventListener('click', () => {
            const anchor = link.getAttribute('href').substr(1);
            gotoAnchor(anchor);
        });
    }
    for (const link of document.querySelectorAll('a[href*=".html"]')) {
        link.addEventListener('click', () => {
            const uri = link.getAttribute('href');
            vscode.postMessage({
                command: 'goto',
                uri: uri,
            });
        });
    }
    window.addEventListener('message', event => {
        const message = event.data;
        switch (message.command) {
            case 'goto':
                gotoAnchor(message.anchor);
                break;
            case 'scrollTo':
                window.scrollTo(message.left, message.top);
                saveState();
                break;
        }
    });
    window.setInterval(() => {
        saveState();
    }, 100);
    saveState();
</SCRIPT>
$1
`);
    fs.readdirSync(srcPath).forEach(function (name) {
        const file = path.join(srcPath, name);
        const stat = fs.statSync(file);
        if (stat && stat.isFile()) {
            if (".html" != path.extname(file)) {
                const uri = currentPanel.webview.asWebviewUri(vscode.Uri.file(path.join(dstPath, name)));
                html = html.replace(name, uri);
            }
        }
    });
    fs.writeFileSync(path.join(dstPath, name), html);
}

function compile(workPath, language, version, srcPath, dstPath) {
    fs.mkdirSync(dstPath, { recursive: true });
    fs.readdirSync(srcPath).forEach(function (name) {
        const file = path.join(srcPath, name);
        const stat = fs.statSync(file);
        if (!stat || !stat.isFile()) {
            return;
        }
        const extname = path.extname(file);
        if (".html" == extname) {
            compileHtml(srcPath, dstPath, language, version, name);
        }
        else if (".css" == extname) {
            compileCss(path.join(workPath, 'doc', 'en-us', '54'), dstPath, name);
        }
        else {
            compileOther(srcPath, dstPath, name);
        }
    });
}

function needCompile(workPath, dstPath) {
    const cfg = path.join(dstPath, '.compiled');
    if (!fs.existsSync(cfg)) {
        return true;
    }
    return (workPath != fs.readFileSync(cfg, 'utf8'));
}

function checkAndCompile(workPath, language, version) {
    const srcPath = path.join(workPath, 'doc', language, version);
    const dstPath = path.join(workPath, 'out', language, version);
    if (needCompile(workPath, dstPath)) {
        if (!fs.existsSync(srcPath)) {
            currentPanel.title = 'Error';
            currentPanel.webview.html = `
<!DOCTYPE html>
<html lang="en">
    <head></head>
    <body>
        <h1>Not Found doc/${language}/${version}/</h1>
    </body>
</html>`;
            return false;
        }
        compile(workPath, language, version, srcPath, dstPath);
        fs.writeFileSync(path.join(dstPath, '.compiled'), workPath);
    }
    currentPanel._language = language;
    currentPanel._version = version;
    return true
}

function openHtml(workPath, file) {
    const language = currentPanel._language;
    const version = currentPanel._version;
    const htmlPath = path.join(workPath, 'out', language, version, file);
    if (currentPanel._file == htmlPath) {
        return;
    }
    currentPanel._file = htmlPath;
    if (!fs.existsSync(htmlPath)) {
        currentPanel.title = 'Error';
        currentPanel.webview.html = `
<!DOCTYPE html>
<html lang="en">
<head></head>
<body>
    <h1>Not Found doc/${language}/${version}/${file}</h1>
</body>
</html>`;
        return;
    }
    const html = fs.readFileSync(htmlPath, 'utf8');
    currentPanel.title = html.match(/<title>(.*?)<\/title>/i)[1];
    currentPanel.webview.html = html;
}

function gotoAnchor(anchor) {
    currentPanel.webview.postMessage({ command: 'goto', anchor: anchor });
}

function parseUri(uri) {
    const l = uri.split(/[\/#]/g);
    return {
        language: l[0],
        version: l[1],
        file: l[2],
        anchor: l[3],
    };
}

function getViewColumn(reveal) {
    if (vscode.window.activeTextEditor) {
        if (vscode.window.activeTextEditor.viewColumn == vscode.ViewColumn.One) {
            return vscode.ViewColumn.Two;
        }
        return vscode.window.activeTextEditor.viewColumn;
    }
    if (reveal) {
        return undefined;
    }
    return vscode.ViewColumn.One;
}

function createPanel(workPath, disposables, viewType) {
    const options = {
        enableScripts: true,
        enableFindWidget: true,
        retainContextWhenHidden: true,
    };
    let panel = vscode.window.createWebviewPanel(viewType, '', { viewColumn: getViewColumn(false), preserveFocus: true }, options);
    panel.webview.onDidReceiveMessage(
        message => {
            switch (message.command) {
                case 'goto':
                    const uri = message.uri.split("#");
                    openHtml(workPath, uri[0]);
                    if (uri[1]) {
                        gotoAnchor(uri[1]);
                    }
                    return;
            }
        },
        null,
        disposables
    );
    panel.onDidDispose(
        () => {
            currentPanel = undefined;
        },
        null,
        disposables
    );
    return panel;
}

function createWebviewPanel(workPath, disposables, viewType, uri) {
    if (currentPanel) {
        try {
            currentPanel.reveal(getViewColumn(true), true);
        } catch (error) {
            currentPanel = undefined;
        }
    }
    if (!currentPanel) {
        currentPanel = createPanel(workPath, disposables, viewType);
    }
    const args = parseUri(uri);
    if (!checkAndCompile(workPath, args.language, args.version)) {
        return;
    }
    openHtml(workPath, args.file);
    if (args.anchor) {
        gotoAnchor(args.anchor);
    }
}

function revealWebviewPanel(workPath, webviewPanel, state) {
    if (!state) {
        webviewPanel.dispose();
        return;
    }
    currentPanel = webviewPanel;
    if (!checkAndCompile(workPath, state.language, state.version)) {
        return;
    }
    openHtml(workPath, state.file);
    currentPanel.webview.postMessage({
        command: 'scrollTo',
        top: state.scrollTop,
        left: state.scrollLeft,
    });
}

function activateLuaDoc(workPath, disposables, LuaDoc) {
    disposables.push(vscode.commands.registerCommand(LuaDoc.OpenCommand, (uri) => {
        try {
            createWebviewPanel(workPath, disposables, LuaDoc.ViewType, uri || "en-us/54/readme.html");
        } catch (error) {
            console.error(error)
        }
    }));

    vscode.window.registerWebviewPanelSerializer(LuaDoc.ViewType, {
        deserializeWebviewPanel(webviewPanel, state) {
            try {
                revealWebviewPanel(workPath, webviewPanel, state)
            } catch (error) {
                console.error(error)
            }
        }
    });
}

function activate(context) {
    activateLuaDoc(context.extensionPath, context.subscriptions, {
        ViewType: context.ViewType || 'test-lua-doc',
        OpenCommand: context.OpenCommand || 'test.lua.doc',
    });
}

exports.activate = activate;
