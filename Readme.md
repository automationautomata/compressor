## Compression console tool

Compress multiple files or one directory to one file and decompress it to destination dir.

Currently only Huffman compression is available. For this type, you can set the number of bytes in the symbol of the alphabet.

How to use:

    ``compressor compress /path/to/file -dest=/path/to/dir``

    ``compressor uncompress /path/to/file -dest=/path/to/dir``

    ``compressor matadata /path/to/file``

The metadata command prints a list of compressed files with their sizes and checksums

