const {spawn} = require('child_process');

const child = spawn('dist/update-go');
child.stdout.pipe(process.stdout);
child.stderr.pipe(process.stderr);
child.on('close', process.exit);
