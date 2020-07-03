package common

// RPC's:
// - start a game
// - Ask for a card
// - play a pair
// - game status (in progress, your turn, game over)


// Start game:
// players call server to join a game
// if a game hasnt been created, server creates a new one
// if it has, player joins the game
// games can only have between 2-7 players
// first player to join is the first to play (identified by indices 0,1,2...)
