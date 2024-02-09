
// Board has a lot of Go-specific (as in the game Go) code in it
// Othello is sort of bolted on that as a best effort

import {$, $$, drawText} from './util.js';
import {noFillFn, neverFillFn, Board} from './board.js';
import {Point, Edge, Polygon, randomEdgePoint} from './primitives.js';

let board = null;
let me = null;
let boardName = null;
let key = null;
let legalMoves = [];
   
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
            pts[i] = {X: board.points[i].x, Y: board.points[i].y, Id: i, Player: 0};
        } else if (board.points[i].player == 'white') {
            pts[i] = {X: board.points[i].x, Y: board.points[i].y, Id: i, Player: 1};
        } else {
            pts[i] = {X: board.points[i].x, Y: board.points[i].y, Id: i, Player: -1};
        }
    }
    return pts;
}

function getMove(pts1, pts2) {
    for (let i=0; i<pts1.length; i++) {
        if (pts1[i].Player == -1 && pts2[i].Player != -1) {
            return i;
        }
    }
    return -1;
}

function getScores(board) {
    const pts = transformPoints(board);
    const scores = [0,0];
    for (let i=0; i<pts.length; i++) {
        if (pts[i].Player == 0) {
            scores[0]++;
        } else if (pts[i].Player == 1) {
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
                legalMoves = json.LegalMoves;
                break;
            case 'Move':
                console.log(json);
                const player = json.Player;
                const points = json.Points;
                legalMoves = json.LegalMoves;
                board.points.forEach((pt, i) => {
                   if (points[i].Player != -1) {
                       pt.player = points[i].Player == 0 ? "black" : "white";
                   } else {
                       pt.player = null;
                   }
                });
                board.player = player == 0 ? "white" : "black";
                board.repaint();
                displayScores(board);
                break;
            case 'JoinGame': {
                key = json.Key; 
                me = 1;
                const boardPlan = JSON.parse(json.BoardPlan);
                const points = json.Points;
                legalMoves = json.LegalMoves;
                board = new Board(canvas);
                initBoard(board, boardPlan);
                board.points.forEach((pt, i) => {
                    if (points[i].Player != -1) {
                        pt.player = points[i].Player == 0 ? "black" : "white";
                    }
                });
                board.player = getNumPieces(board) % 2 == 0 ? "black" : "white";
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
        legalMoves = [];
    });

    $('#canvas').addEventListener('mousemove', (e) => {
        if (board && getNumPieces(board) < 4) {
            board.hover(e.offsetX, e.offsetY);
            board.repaint();
        } else if (board && legalMoves.length > 0) {
            const p = board.hover(e.offsetX, e.offsetY);
            if (p && !legalMoves.includes(p.id)) {
                p.hover = false;
            }
            board.repaint();
        }
    });

    canvas.addEventListener('click', e => {
        if (board && getNumPieces(board) < 4) {
            board.click(e.offsetX, e.offsetY);
            board.repaint();
            if (getNumPieces(board) == 4) {
                drawText(board.canvas.getContext('2d'), 
                    "You may now start a game", 
                    new Point(canvas.width/2, 40),
                    'red', 
                    'bold 28px sans', 
                    true);
                $('#new').disabled = false;
                $('#new-ai').disabled = false;
            }
        } else if (board && legalMoves.length > 0) {
            // Check for legality of move
            const p = board.hover(e.offsetX, e.offsetY);
            if (!p) 
                return;
            if (!legalMoves.includes(p.id)) 
                return;
            const pts = transformPoints(board);
            board.click(e.offsetX, e.offsetY);
            board.repaint();
            const pts2 = transformPoints(board);
            const move = getMove(pts, pts2);
            if (move != -1) {
                // Take back move and wait for server to give us the updated board
                board.points[move].player = null;
                board.player = getTurn(board) % 2 == 0 ? "black" : "white";
                conn.send(JSON.stringify({Action: 'Move', Key: key, Move: move}));    
            }
        }
    });

    $('#new').addEventListener('click', () => {
        const pts = transformPoints(board);
        const req = {Action: 'NewGame', AIGame: false, BoardName: boardName, Points: pts};
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
