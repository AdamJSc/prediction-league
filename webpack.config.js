const parseBuildMode = (args) => {
    for (let idx in args) {
        let arg = args[idx]
        if (arg.startsWith('--mode=')) {
            let parts = arg.split('=', -1)
            if (parts.length === 2) {
                return parts[1]
            }
            return 'development'
        }
    }
}

const buildMode = parseBuildMode(process.argv)
console.log(`building in ${buildMode} mode...`)

const path = require('path')

const VueLoaderPlugin = require('vue-loader/lib/plugin')
const CopyPlugin = require('copy-webpack-plugin')

module.exports = {
    mode: buildMode,
    entry: {
        main: path.resolve(__dirname, 'resources', 'src', 'js', 'main.js')
    },
    output: {
        path: path.resolve(__dirname, 'resources', 'dist'),
        filename: 'app.[name].js'
    },
    module: {
        rules: [
            {
                test: /.vue$/,
                loader: 'vue-loader',
            },
            {
                test: /.js$/,
                exclude: [path.resolve(__dirname, 'node_modules')],
                use: [{
                    loader: 'babel-loader',
                    options: {
                        presets: [
                            ['@babel/preset-env']
                        ]
                    }
                }]
            },
        ],
    },
    plugins: [
        new VueLoaderPlugin(),
        new CopyPlugin({
            patterns: [{
                context: path.resolve(__dirname, 'resources', 'src', 'images'),
                from: '**/*',
                to: 'img',
                force: true
            }]
        }),
    ],
    resolve: {
        extensions: ['.json', '.js', '.jsx'],
    }
}
