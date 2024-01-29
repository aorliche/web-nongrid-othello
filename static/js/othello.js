
// Board has a lot of Go-specific (as in the game Go) code in it
// Othello is sort of bolted on that as a best effort

import {$, $$, drawText} from './util.js';
import {noFillFn, neverFillFn, Board} from './board.js';
import {Point, Edge, Polygon, randomEdgePoint} from './primitives.js';

let board = null;
let me = null;
let boardName = null;
let key = null;
   
function initBoard(board, boardPlan) {
    const fn = (typ, n) => {
        if (n == -1) {
            return neverFillFn;
        } else if (n == 0) {
            return noFillFn;
        } else if (typ == 'fill') {
            return (a, b) => board.fill(a,n,b);
        } else {
            return (a, b) => board.placeOne(a,n,b);
        }
    }
    boardPlan.forEach(round => {
        const arr = [];
        for (let i=0; i<round.sav.length; i++) {
            arr.push(fn(round.typ, round.sav[i].n));
        }
        board.loop(arr);
    });
    board.initNeighbors();
    board.repaint();
}

function getNumPieces(board) {
    let n=0;
    for (let i=0; i<board.points.length; i++) {
        // black or white
        if (board.points[i].player) {
            n++;
        }
    }
    return n;
}

function getTurn(board) {
    const numPieces = getNumPieces(board);
    if (numPieces < 4) {
        return -1;
    }
    return numPieces - 4;
}

function transformPoints(board) {
    const pts = [];
    for (let i=0; i<board.points.length; i++) {
        if (board.points[i].player == 'black') {
            pts[i] = 0;
        } else if (board.points[i].player == 'white') {
            pts[i] = 1;
        } else {
            pts[i] = -1;
        }
    }
    return pts;
}

// For whatever reason, board.neighbors is an object and not an Array
function transformNeighbors(board) {
    const ns = [];
    for (let i=0; i<board.points.length; i++) {
        ns[i] = board.neighbors[i];
    }
    return ns;
}

function getShortestPaths(board, p1, p2) {
    function node2path(n, path) {
        while (n.prev != null) {
            path.push(n.cur);
            n = n.prev;
        }
        path.push(n.cur);
        path.reverse();
    }
    const visited = [];
    const pts = transformPoints(board);
    for (let i=0; i<pts.length; i++) {
        visited.push(false);
    }
    visited[p1] = true; 
    const start = {prev: null, cur: p1};
    let frontier = [start];
    while (frontier.length > 0) {
        const next = [];
        const nextVisited = [];
        let finished = false;
        for (let i=0; i<frontier.length; i++) {
            const ns = board.neighbors[frontier[i].cur];
            for (let j=0; j<ns.length; j++) {
                const p = ns[j];
                if (visited[p]) {
                    continue
                }
                nextVisited.push(p);
                next.push({prev: frontier[i], cur: p});
                if (p == p2) {
                    finished = true;
                }
            }
        }
        if (finished) {
            const paths = [];
            next.forEach(node => {
                if (node.cur == p2) {
                    const path = [];
                    node2path(node, path);
                    paths.push(path);
                }
            });
            return paths;
        }
        frontier = next;
        nextVisited.forEach(p => {
            visited[p] = true;
        });
    }
    return [];
}

function getPossibleMoves(board) {
    const me = getTurn(board) % 2;
    const other = 1-me;
    const from = [];  
    const to = [];
    const pts = transformPoints(board);
    pts.forEach((player, i) => {
        if (player == me) {
            from.push(i);
        } else if (player == -1) {
            for (let j=0; j<board.neighbors[i].length; j++) {
                const np = board.neighbors[i][j];
                if (pts[np] == other) {
                    to.push(i);
                    break
                }
            }
        }
    });
    const moves = [];
    to.forEach(toP => {
        from.forEach(fromP => {
            const paths = getShortestPaths(board, fromP, toP);
            let valid = false;
            nextpath: 
            for (let i=0; i<paths.length; i++) {
                if (paths[i].length < 3) {
                    continue;
                }
                for (let j=1; j<paths[i].length - 1; j++) {
                    if (pts[paths[i][j]] != other) {
                        continue nextpath;
                    }
                }
                valid = true;
                break;
            }
            if (valid) {
                moves.push([fromP, toP]);
            }
        });
    });
    return moves;
}

function getMove(pts1, pts2) {
    for (let i=0; i<pts1.length; i++) {
        if (pts1[i] == -1 && pts2[i] != -1) {
            return i;
        }
    }
    return -1;
}

function getScores(board) {
    const pts = transformPoints(board);
    const scores = [0,0];
    for (let i=0; i<pts.length; i++) {
        if (pts[i] == 0) {
            scores[0]++;
        } else if (pts[i] == 1) {
            scores[1]++;
        }
    }
    return scores;
}

function displayScores(board) {
    const scores = getScores(board);
    $('#black').innerText = scores[0];
    $('#white').innerText = scores[1];
}

function gameOver(board) {

}

window.addEventListener('load', () => {
    const canvas = $('#canvas');
    const conn = new WebSocket(`ws://${location.host}/ws`);
    
    conn.onmessage = e => {
        const json = JSON.parse(e.data);
        switch (json.Action) {
            case 'ListBoards':
                const boardNames = json.BoardNames;
                boardNames.forEach(name => {
                    const existing = $$('#boards option');
                    let found = false;
                    for (let i=0; i<existing.length; i++) {
                        if (existing[i].innerText == name) {
                            found = true;
                            break;
                        }
                    }
                    if (!found) {
                        const opt = document.createElement('option');
                        opt.innerText = name;
                        $('#boards').appendChild(opt);
                    }
                });
                break;
            case 'ListGames':
                const keys = json.Keys;
                keys.sort((a,b) => a-b);

                const select = $('select[name="games-list"]');
                const toAdd = [];
                const games = [...select.options].map(opt => parseInt(opt.value));
                if (key !== null && games.includes(key)) {
                    for (let i=0; i<select.options.length; i++) {
                        const opt = select.options[i];
                        if (parseInt(opt.value) == key) {
                            select.remove(i);
                            break;
                        }
                    }
                }
                for (let i=0; i<select.options.length; i++) {
                    const opt = select.options[i];
                    if (!keys.includes(parseInt(opt.value))) {
                        select.remove(i--);
                    }
                }
                keys.forEach(k => {
                    if (!games.includes(k) && k != key) {
                        const opt = document.createElement('option');
                        opt.value = k;
                        opt.innerHTML = `Game ${k}`;
                        select.appendChild(opt);
                    }
                });
                break;
            case 'LoadBoard': {
                const boardPlan = JSON.parse(json.BoardPlan);
                board = new Board(canvas);
                initBoard(board, boardPlan);
                break;
            }
            case 'NewGame':
                key = json.Key;
                me = 0;
                displayScores(board);
                break;
            case 'Move':
                const move = json.Move;
                const player = json.Player;
                // We check the possible moves
                // And perform the move at the same time
                const moves = getPossibleMoves(board);
                let paths = [];
                let valid = false;
                for (let i=0; i<moves.length; i++) {
                    if (moves[i][1] == move) {
                        valid = true;
                        const ps = getShortestPaths(board, moves[i][0], move);
                        paths = paths.concat(ps);
                    }
                }
                // The move wasn't in the possible moves
                if (!valid) {
                    break;
                }
                paths.sort((a,b) => a.length - b.length);
                function validPath(path) {
                    const pts = transformPoints(board);
                    for (let i=1; i<path.length-1; i++) {
                        if (pts[path[i]] != 1-pts[path[0]]) {
                            return false;
                        }
                    }
                    return true;
                }
                paths.forEach(path => {
                    if (!validPath(path)) {
                        return;
                    }
                    for (let j=0; j<path.length; j++) {
                        board.points[path[j]].player = player == 0 ? "black" : "white";
                    }
                });
                //board.points[move].player = player == 0 ? "black" : "white";
                board.player = player == 0 ? "white" : "black";
                board.repaint();
                displayScores(board);
                break;
            case 'JoinGame': {
                key = json.Key; 
                me = 1;
                const boardPlan = JSON.parse(json.BoardPlan);
                const pts = json.Points;
                board = new Board(canvas);
                initBoard(board, boardPlan);
                let n = 0;
                for (let i=0; i<pts.length; i++) {
                    if (pts[i] != -1) {
                        n++;
                        board.points[i].player = pts[i] == 0 ? "black" : "white";
                    } else {
                        board.points[i].player = null;
                    }
                }
                board.player = n % 2 == 0 ? "black" : "white";
                board.repaint();
                displayScores(board);
                break;
            }
        }
    }
    
    conn.onopen = e => {
        conn.send(JSON.stringify({Action: 'ListBoards'}));
    }
    
    setInterval(e => {
        if (!conn.readyState == 1) return;
        conn.send(JSON.stringify({Action: 'ListGames'}));
    }, 1000);

    $('#load').addEventListener('click', () => {
        const idx = $('#boards').selectedIndex;
        if (idx == -1) return;
        boardName = $('#boards').options[idx].innerText;
        const req = {Action: 'LoadBoard', BoardName: boardName};
        conn.send(JSON.stringify(req));
    });

    $('#canvas').addEventListener('mousemove', (e) => {
        if (board && (getNumPieces(board) < 4 || (!gameOver(board) && (me == getTurn(board) % 2)))) {
            board.hover(e.offsetX, e.offsetY);
            board.repaint();
        }
    });

    canvas.addEventListener('click', e => {
        if (board && (getNumPieces(board) < 4 || (!gameOver(board) && (me == getTurn(board) % 2)))) {
            const pts = transformPoints(board);
            board.click(e.offsetX, e.offsetY);
            board.repaint();
            if (getNumPieces(board) == 4) {
                const moves = getPossibleMoves(board);
                if (moves.length == 0) {
                    drawText(board.canvas.getContext('2d'), 
                        "There are no possible starting moves, try again", 
                        new Point(canvas.width/2, 40),
                        'red', 
                        'bold 28px sans', 
                        true);
                    drawText(board.canvas.getContext('2d'), 
                        "(place pieces next to each other)", 
                        new Point(canvas.width/2, 70),
                        'red', 
                        'bold 28px sans', 
                        true);
                } else {
                    drawText(board.canvas.getContext('2d'), 
                        "You may now start a game", 
                        new Point(canvas.width/2, 40),
                        'red', 
                        'bold 28px sans', 
                        true);
                    $('#new').disabled = false;
                    $('#new-ai').disabled = false;
                }
            }
            // Playing a move
            // We take back a move and only put it back after an ack from the server
            if (getNumPieces(board) > 4) {
                const pts2 = transformPoints(board);
                const move = getMove(pts, pts2);
                if (move != -1) {
                    board.points[move].player = null;
                    board.player = getTurn(board) % 2 == 0 ? "black" : "white";
                    conn.send(JSON.stringify({Action: 'Move', Key: key, Move: move}));    
                }
            }
        }
    });

    $('#new').addEventListener('click', () => {
        const pts = transformPoints(board);
        const ns = transformNeighbors(board);
        const req = {Action: 'NewGame', AIGame: false, BoardName: boardName, Points: pts, Neighbors: ns};
        conn.send(JSON.stringify(req));
        $('#new').disabled = true;
        $('#new-ai').disabled = true;
    });

    $('#join').addEventListener('click', () => {
        const idx = $('select[name="games-list"]').selectedIndex;
        if (idx == -1) return;
        const k = parseInt($('select[name="games-list"]').options[idx].value);
        const req = {Action: 'JoinGame', Key: k};
        conn.send(JSON.stringify(req));
    });

});
