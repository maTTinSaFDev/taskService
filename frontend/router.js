export function readRoute() {
    const match = window.location.pathname.match(/^\/boards\/(\d+)\/?$/);
    const requestedBoardId = match ? Number(match[1]) : 0;
    const boardId = Number.isInteger(requestedBoardId) && requestedBoardId > 0 ? requestedBoardId : null;
    return {boardId};
}

export function boardURL(boardId = null) {
    const path = boardId ? `/boards/${boardId}` : "/boards";
    return path;
}

export function setBoardRoute(boardId, replace = false) {
    const url = boardURL(boardId);
    if (`${window.location.pathname}${window.location.search}` === url) return;
    window.history[replace ? "replaceState" : "pushState"]({boardId}, "", url);
}

export function onRouteChange(handler) {
    window.addEventListener("popstate", handler);
}
