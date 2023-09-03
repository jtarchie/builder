import Fuse from './fuse.js';
// Create the Fuse index
const myIndex = Fuse.createIndex(["id", "title", "contents"], documents)
// Serialize and save it
JSON.stringify(myIndex.toJSON())