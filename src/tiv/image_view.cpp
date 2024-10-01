// Copyright 2023 ljrrjl. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

#include <algorithm>  // for min
#include <memory>     // for make_shared
#include <string>     // for string, wstring
#include <utility>    // for move
#include <vector>     // for vector
#include <sstream>
#include <fstream>

#include "ftxui/dom/deprecated.hpp"   // for text, vtext
#include "ftxui/dom/elements.hpp"     // for Element, text, vtext
#include "ftxui/dom/node.hpp"         // for Node
#include "ftxui/dom/requirement.hpp"  // for Requirement
#include "ftxui/screen/box.hpp"       // for Box
#include "ftxui/screen/screen.hpp"    // for Pixel, Screen
#include "ftxui/screen/string.hpp"  // for string_width, Utf8ToGlyphs, to_string

#define cimg_display 0
#include "CImg.h"
#include "tiv_lib.h"

namespace ftxui {

namespace {

using ftxui::Screen;

class ImageView: public Node {
 public:
  explicit ImageView(std::string_view url) : url_(url) {}

  void ComputeRequirement() override {
    if(!is_loaded_) {
      img_ = tiv::load_rgb_CImg(url_.c_str());
      is_loaded_ = true;
    }

    requirement_.min_x = img_.width() / 4;
    requirement_.min_y = img_.height() / 8;
  }

  void Render(Screen& screen) override {

    // Create a copy of the image to work with
    // It's slow on the render cycle, yeah, but it's a quick and dirty solution
    cimg_library::CImg<unsigned char> working_img(img_); 

    auto get_pixel = [this, &working_img](int row, int col) -> unsigned long {
      return (((unsigned long) working_img(row, col, 0, 0)) << 16) 
          | (((unsigned long) working_img(row, col, 0, 1)) << 8)
          | (((unsigned long) working_img(row, col, 0, 2)));
      };
    auto origin_image_width = (box_.x_max - box_.x_min + 1) * 4;
    auto origin_image_height = (box_.y_max - box_.y_min + 1) * 8;
    auto new_size = tiv::size(working_img).fitted_within(tiv::size(origin_image_width, origin_image_height));
    working_img.resize(new_size.width, new_size.height, -100, -100, 5);


    auto flags = tiv::FLAG_MODE_256;
    tiv::CharData lastCharData;
    auto screen_y = box_.y_min;
    for (int y = 0; y <= working_img.height() - 8; y += 8) {
        auto screen_x = box_.x_min;
        for (int x = 0; x <= working_img.width() - 4; x += 4) {
            if(screen_x > box_.x_max)
              break;
            std::stringstream output;
            tiv::CharData charData = tiv::findCharData(get_pixel, x, y, tiv::FLAG_MODE_256);
            ftxui::Color bgColor(charData.bgColor[0], charData.bgColor[1],charData.bgColor[2]);
            ftxui::Color fgColor(charData.fgColor[0], charData.fgColor[1],charData.fgColor[2]);
            tiv::printCodepoint(output, charData.codePoint);
            auto pixel = ftxui::Pixel();
            pixel.background_color = bgColor;
            pixel.foreground_color = fgColor;
            pixel.character = output.str();
            screen.PixelAt(screen_x++, screen_y) = pixel;
        }
        ++screen_y;
    }
  }

 private:
  std::string url_;
  
  cimg_library::CImg<unsigned char> img_;
  int         width_;
  int         height_;
  bool       is_loaded_;
};

}  // namespace

Element image_view(std::string_view url) {
  return std::make_shared<ImageView>(url);
}

}  // namespace ftxui