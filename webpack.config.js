const buildMode = process.argv.mode
console.log(`building in ${buildMode} mode...`)

const path = require('path')

const VueLoaderPlugin = require('vue-loader/lib/plugin')

module.exports = {
    mode: buildMode,
    entry: path.resolve(__dirname, 'resources', 'src', 'js', 'main.js'),
    output: {
        path: path.resolve(__dirname, 'resources', 'dist'),
        filename: 'app.js'
    },
    module: {
        rules: [
            {
                test: /.vue$/,
                loader: "vue-loader",
            },
            {
                test: /.js$/,
                exclude: [path.resolve(__dirname, "node_modules")],
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
        new VueLoaderPlugin()
    ],
    resolve: {
        extensions: [".json", ".js", ".jsx"],
    },
}
