
// Board has a lot of Go-specific (as in the game Go) code in it
// Othello is sort of bolted on that as a best effort

import {$, $$, drawText} from './util.js';
import {noFillFn, neverFillFn, Board} from './board.js';
import {Point, Edge, Polygon, randomEdgePoint} from './primitives.js';

let board = null;
   
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
    numPieces = getNumPieces(board);
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
            console.log("here");
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

// Turn determines player
// Candidates are empty spaces next to other player's pieces
/*func (board *Board) GetPossibleMoves() [][2]int {
    me := board.Turn % 2
    other := 1-me 
    from := []int{}
    to := []int{}
    for p,player := range board.Points {
        if player == me {
            from = append(from, p)
        } else if player == -1 {
            for _,np := range board.Neighbors[p] {
                if board.Points[np] == other {
                    to = append(to, p)
                    break
                }
            }
        }
    }
    moves := [][2]int{}
    for _,toP := range to {
        for _,fromP := range from {
            paths := board.GetShortestPaths(fromP, toP)
            // Only those paths with all the other player's pieces are allowed
            // Other than starting and ending points
            // Also they must have length > 3
            valid := false
            nextpath:
            for _,path := range paths {
                if len(path) < 3 {
                    continue
                }
                for i := 1; i < len(path) - 1; i++ {
                    if board.Points[path[i]] != other {
                        continue nextpath
                    }
                }
                valid = true
                break
            }
            if valid {
                moves = append(moves, [2]int{fromP, toP})
            }
        }
    }
    return moves
}*/

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
                break;
            case 'LoadBoard':
                const boardPlan = JSON.parse(json.BoardPlan);
                board = new Board(canvas);
                initBoard(board, boardPlan);
                break;
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
        const req = {Action: 'LoadBoard', BoardName: $('#boards').options[idx].innerText};
        conn.send(JSON.stringify(req));
    });

    $('#canvas').addEventListener('mousemove', (e) => {
        if (board && getNumPieces(board) < 4) {
            board.hover(e.offsetX, e.offsetY);
            board.repaint();
        }
    });

    canvas.addEventListener('click', e => {
        if (board && getNumPieces(board) < 4) {
            board.click(e.offsetX, e.offsetY);
            board.repaint();
            const pts = transformPoints(board);
            for (let i=0; i<pts.length; i++) {
                for (let j=i+1; j<pts.length; j++) {
                    if (pts[i] == pts[j] && pts[i] != -1) {
                        console.log(i, j, pts[i], pts[j], getShortestPaths(board, i, j));
                    }
                }
            }
        }
    });

});

        /*
        // List of integer game ids
        json.sort((a,b) => a-b);

        const select = $('select[name="games-list"]');
        const toAdd = [];
        const games = [...select.options].map(opt => parseInt(opt.value));
        if (game && games.includes(game.id)) {
            for (let i=0; i<select.options.length; i++) {
                const opt = select.options[i];
                if (parseInt(opt.value) == game.id) {
                    select.remove(i);
                    break;
                }
            }
        } 
        for (let i=0; i<select.options.length; i++) {
            const opt = select.options[i];
            if (!json.includes(parseInt(opt.value))) {
                select.remove(i--);
            }
        }
        json.forEach(key => {
            if (!games.includes(key) && !(game && game.id == key)) {
                const opt = document.createElement('option');
                opt.value = key;
                opt.innerHTML = `Game ${key}`;
                select.appendChild(opt);
            }
        });
    }
    
    const connBoards = new WebSocket(`ws://${location.host}/boards`);

    setInterval(e => {
        if (!conn.readyState == 1) return;
        conn.send(JSON.stringify({Action: 'List'}));
    }, 1000);
    
    setInterval(e => {
        if (!connBoards.readyState == 1) return;
        connBoards.send(JSON.stringify({Action: 'List'}));
    }, 1000);

    $('#load').addEventListener('click', () => {
        const idx = $('#boards').selectedIndex;
        if (idx == -1) return;
        const req = {Action: 'Load', Player: $('#boards').options[idx].innerText};
        connBoards.send(JSON.stringify(req));
    });

    connBoards.onmessage = e => {
        const msg = JSON.parse(e.data);
        if (msg.Action == 'List') {
            JSON.parse(msg.Payload).forEach(name => {
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
        } else if (msg.Action == 'Load') {
            boardjson = msg.Payload;
            initBoard(new Board(canvas));
        }
    }*/
