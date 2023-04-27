# Distributed under the OSI-approved BSD 3-Clause License.  See accompanying
# file Copyright.txt or https://cmake.org/licensing for details.

cmake_minimum_required(VERSION 3.22.4)

# We name the project and the target for the ExternalProject_Add() call
# to something that will highlight to the user what we are working on if
# something goes wrong and an error message is produced.

project(magic_enum-populate NONE)



include(ExternalProject)
ExternalProject_Add(magic_enum-populate
                     "UPDATE_DISCONNECTED" "False" "URL" "https://github.com/Neargye/magic_enum/archive/3d1f6a5a2a3fbcba077e00ad0ccc2dd9fefc2ca7.zip"
                    SOURCE_DIR          "/home/luis.antunes.veiga/sdb/MDS-go/GEDS/_deps/magic_enum-src"
                    BINARY_DIR          "/home/luis.antunes.veiga/sdb/MDS-go/GEDS/_deps/magic_enum-build"
                    CONFIGURE_COMMAND   ""
                    BUILD_COMMAND       ""
                    INSTALL_COMMAND     ""
                    TEST_COMMAND        ""
                    USES_TERMINAL_DOWNLOAD  YES
                    USES_TERMINAL_UPDATE    YES
)


