cmake_minimum_required(VERSION 3.20)

project(GoALLCStatepoints LANGUAGES C CXX)

find_package(LLVM REQUIRED CONFIG)
if(NOT LLVM_VERSION_MAJOR EQUAL 23)
  message(FATAL_ERROR
    "GoALLCStatepoints requires LLVM 23, found ${LLVM_PACKAGE_VERSION}")
endif()
if(WIN32 OR CYGWIN)
  message(FATAL_ERROR "GoALLCStatepoints does not support Windows plugins")
endif()

list(APPEND CMAKE_MODULE_PATH "${LLVM_CMAKE_DIR}")
include(AddLLVM)

add_definitions(${LLVM_DEFINITIONS})
include_directories(SYSTEM ${LLVM_INCLUDE_DIRS})

add_llvm_pass_plugin(GoALLCStatepoints
  GoALLCPreCodeGen.cpp
  GoALLCStackMapPrinter.cpp
  GoALLCStatepoints.cpp
  GoALLCStatepointPlugin.cpp
)

set_target_properties(GoALLCStatepoints PROPERTIES
  CXX_STANDARD 17
  CXX_STANDARD_REQUIRED YES
  CXX_EXTENSIONS NO
)
if(APPLE)
  target_link_options(GoALLCStatepoints PRIVATE
    "LINKER:-undefined,dynamic_lookup"
  )
endif()

include(CTest)
if(BUILD_TESTING)
  find_program(GOALLC_LLC_EXECUTABLE
    NAMES llc
    HINTS "${LLVM_TOOLS_BINARY_DIR}"
    REQUIRED
  )
  add_test(
    NAME GoALLCStatepoints.Load
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-report
      -filetype=null
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/load.ll"
  )
  set_tests_properties(GoALLCStatepoints.Load PROPERTIES
    PASS_REGULAR_EXPRESSION
      "GoALLCStatepoints: ran pre-codegen pipeline for"
  )

  add_test(
    NAME GoALLCStatepoints.Rewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/statepoint.ll"
  )
  set_tests_properties(GoALLCStatepoints.Rewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "slot.relocated"
  )

  add_test(
    NAME GoALLCStatepoints.AllocaDerivedRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/alloca-derived.ll"
  )
  set_tests_properties(GoALLCStatepoints.AllocaDerivedRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "selected.relocated"
  )

  add_test(
    NAME GoALLCStatepoints.GoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -statepoint-callsite-position=call-start
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/statepoint.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/statepoint.ll"
  )

  add_test(
    NAME GoALLCStatepoints.AllocaDerivedGoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -statepoint-callsite-position=call-start
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/alloca-derived.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/alloca-derived.ll"
  )

  add_test(
    NAME GoALLCStatepoints.LiveAggregateFails
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/live-aggregate.ll"
  )
  set_tests_properties(GoALLCStatepoints.LiveAggregateFails PROPERTIES
    WILL_FAIL TRUE
  )

  add_test(
    NAME GoALLCStatepoints.ConditionalRelocationFails
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.ConditionalRelocationFails PROPERTIES
    WILL_FAIL TRUE
  )

  add_test(
    NAME GoALLCStatepoints.InvokeFails
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/unsupported-invoke.ll"
  )
  set_tests_properties(GoALLCStatepoints.InvokeFails PROPERTIES
    WILL_FAIL TRUE
  )

  add_test(
    NAME GoALLCStatepoints.ParamAttrFails
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/unsupported-param-attr.ll"
  )
  set_tests_properties(GoALLCStatepoints.ParamAttrFails PROPERTIES
    WILL_FAIL TRUE
  )
endif()
