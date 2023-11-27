const { spawn } = require('child_process');

//commands
const runNetworkCmd = spawn('../main', ['run -p 10000'], {shell: true});

document.getElementById('runNetworkBtn').addEventListener('click', executeRunNetwork);
const logger = document.getElementById('console');

async function executeRunNetwork() {
    try {
        const result = await runNetwork();
        logger.innerText = logger.innerText +`${result}`;
    } catch (error) {
        logger.innerText = logger.innerText +`${error}`;
    }
}

function runNetwork() {
    return new Promise((resolve, reject) => {
        runNetworkCmd.stdout.on('data', data => {
            resolve(`\n${data}`);
        });

        runNetworkCmd.stderr.on('data', data => {
            resolve(`\n${data}`);
        });

        runNetworkCmd.on('error', (error) => {
            resolve(`\n${error.message}`);
        });
    });
}
