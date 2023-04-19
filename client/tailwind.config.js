/** @type {import('tailwindcss').Config} */
const { blackA, violet } = require('@radix-ui/colors');

module.exports = {
    content: [
        "./app/**/*.{js,ts,jsx,tsx,mdx}",
        "./lib/**/*.tsx"
    ],
    theme: {
        extend: {
            colors: {
                ...blackA,
                ...violet,
            },
        },
    },
    plugins: [],
}

