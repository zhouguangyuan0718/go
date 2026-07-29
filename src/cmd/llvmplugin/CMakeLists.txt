cmake_minimum_required(VERSION 3.20)

project(GoALLCStatepoints LANGUAGES C CXX)

set(GOALLC_OBJVIEW_EXECUTABLE "" CACHE FILEPATH
  "optional objview executable for structured GoObj checks")

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
    NAME GoALLCStatepoints.FrontendMarkersPreserved
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/statepoint.ll"
  )
  set_tests_properties(GoALLCStatepoints.FrontendMarkersPreserved PROPERTIES
    PASS_REGULAR_EXPRESSION
      "go-stack-growth-statepoint"
  )

  add_test(
    NAME GoALLCStatepoints.GCLeafMarkers
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/gc-leaf-markers.ll"
  )
  set_tests_properties(GoALLCStatepoints.GCLeafMarkers PROPERTIES
    PASS_REGULAR_EXPRESSION
      "callsite_leaf"
    FAIL_REGULAR_EXPRESSION
      "gc\\.statepoint"
  )

  add_test(
    NAME GoALLCStatepoints.NestParamAttrRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/nest-param-attr.ll"
  )
  set_tests_properties(GoALLCStatepoints.NestParamAttrRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "ptr nest %context"
  )

  add_test(
    NAME GoALLCStatepoints.IndirectCalleeConservativeRoot
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/indirect-callee.ll"
  )
  set_tests_properties(GoALLCStatepoints.IndirectCalleeConservativeRoot PROPERTIES
    PASS_REGULAR_EXPRESSION "gc-live"
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
      -verify-machineinstrs
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/statepoint.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/statepoint.ll"
  )

  add_test(
    NAME GoALLCStatepoints.AllocaDerivedGoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -verify-machineinstrs
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/alloca-derived.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/alloca-derived.ll"
  )

  add_test(
    NAME GoALLCStatepoints.AArch64FrameGoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -verify-machineinstrs
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/aarch64-frame.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/aarch64-frame.ll"
  )

  add_test(
    NAME GoALLCStatepoints.MultipleCallsRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/multiple-calls.ll"
  )
  set_tests_properties(GoALLCStatepoints.MultipleCallsRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "q.relocated2"
  )

  add_test(
    NAME GoALLCStatepoints.MultipleCallsGoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -verify-machineinstrs
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/multiple-calls.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/multiple-calls.ll"
  )

  if(GOALLC_OBJVIEW_EXECUTABLE)
    if(NOT EXISTS "${GOALLC_OBJVIEW_EXECUTABLE}")
      message(FATAL_ERROR
        "GOALLC_OBJVIEW_EXECUTABLE does not exist: "
        "${GOALLC_OBJVIEW_EXECUTABLE}")
    endif()
    find_package(Python3 REQUIRED COMPONENTS Interpreter)
    add_test(
      NAME GoALLCStatepoints.MultipleCallsObjView
      COMMAND
        "${Python3_EXECUTABLE}"
        "${CMAKE_CURRENT_SOURCE_DIR}/testdata/check-multiple-calls.py"
        --objview "${GOALLC_OBJVIEW_EXECUTABLE}"
        --object "${CMAKE_CURRENT_BINARY_DIR}/multiple-calls.goobj"
    )
    set_tests_properties(GoALLCStatepoints.MultipleCallsObjView PROPERTIES
      DEPENDS GoALLCStatepoints.MultipleCallsGoObj
    )

    add_test(
      NAME GoALLCStatepoints.AArch64FrameObjView
      COMMAND
        "${Python3_EXECUTABLE}"
        "${CMAKE_CURRENT_SOURCE_DIR}/testdata/check-aarch64-frame.py"
        --objview "${GOALLC_OBJVIEW_EXECUTABLE}"
        --object "${CMAKE_CURRENT_BINARY_DIR}/aarch64-frame.goobj"
    )
    set_tests_properties(GoALLCStatepoints.AArch64FrameObjView PROPERTIES
      DEPENDS GoALLCStatepoints.AArch64FrameGoObj
    )

    add_test(
      NAME GoALLCStatepoints.ConditionalRelocationObjView
      COMMAND
        "${Python3_EXECUTABLE}"
        "${CMAKE_CURRENT_SOURCE_DIR}/testdata/check-conditional-relocation.py"
        --objview "${GOALLC_OBJVIEW_EXECUTABLE}"
        --object "${CMAKE_CURRENT_BINARY_DIR}/conditional-relocation.goobj"
    )
    set_tests_properties(GoALLCStatepoints.ConditionalRelocationObjView PROPERTIES
      DEPENDS GoALLCStatepoints.ConditionalRelocationGoObj
    )
  endif()

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
    NAME GoALLCStatepoints.ConditionalRelocationRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.ConditionalRelocationRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "p\\.relocated\\.merge.* = phi ptr"
  )

  add_test(
    NAME GoALLCStatepoints.ConditionalPhiEdgeUseRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.ConditionalPhiEdgeUseRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "selected = phi ptr \\[ %p\\.relocated, %call \\]"
  )

  add_test(
    NAME GoALLCStatepoints.ConditionalRelocationStableID
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.ConditionalRelocationStableID PROPERTIES
    PASS_REGULAR_EXPRESSION
      "i64 -6495201133029790441"
  )

  add_test(
    NAME GoALLCStatepoints.ConditionalRelocationGoObj
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -verify-machineinstrs
      -filetype=obj
      -o "${CMAKE_CURRENT_BINARY_DIR}/conditional-relocation.goobj"
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/conditional-relocation.ll"
  )

  add_test(
    NAME GoALLCStatepoints.BranchSafepointsRelocationRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/branch-safepoints-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.BranchSafepointsRelocationRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "p\\.relocated\\.merge.* = phi ptr"
  )

  add_test(
    NAME GoALLCStatepoints.SequentialConditionalRelocationRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/sequential-conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.SequentialConditionalRelocationRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "p\\.relocated\\.merge.* = phi ptr"
  )

  add_test(
    NAME GoALLCStatepoints.SequentialConditionalLiveSetRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/sequential-conditional-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.SequentialConditionalLiveSetRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "\"gc-live\"\\(ptr %p\\.relocated\\.merge"
  )

  add_test(
    NAME GoALLCStatepoints.SameBlockRelocationChainRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/statepoint.ll"
  )
  set_tests_properties(GoALLCStatepoints.SameBlockRelocationChainRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "\"gc-live\"\\(ptr %p\\.relocated2"
  )

  add_test(
    NAME GoALLCStatepoints.LoopRelocationRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/loop-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.LoopRelocationRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "p\\.relocated\\.merge\\.1 = phi ptr"
  )

  add_test(
    NAME GoALLCStatepoints.IrreducibleRelocationRewrite
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -goallc-pass-plugin-emit-ir
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/irreducible-relocation.ll"
  )
  set_tests_properties(GoALLCStatepoints.IrreducibleRelocationRewrite PROPERTIES
    PASS_REGULAR_EXPRESSION
      "p\\.relocated\\.merge.* = phi ptr"
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

  add_test(
    NAME GoALLCStatepoints.InvalidGCLeafDefinitionFails
    COMMAND
      "${GOALLC_LLC_EXECUTABLE}"
      "-load-pass-plugin=$<TARGET_FILE:GoALLCStatepoints>"
      -filetype=null
      -o -
      "${CMAKE_CURRENT_SOURCE_DIR}/testdata/invalid-gc-leaf-definition.ll"
  )
  set_tests_properties(GoALLCStatepoints.InvalidGCLeafDefinitionFails PROPERTIES
    WILL_FAIL TRUE
  )
endif()
