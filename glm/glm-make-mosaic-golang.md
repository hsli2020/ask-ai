用golang语言编写一个程序，搜索当前目录下的很多类似logo的小尺寸图像文件，
- 把每12张图片按4行3列的方式排列，拼接成一张大图，
- 每张小图之间留有空隙，
- 小图加上淡色边框，
- 保持小图的比例的前提下，缩放小图的尺寸占满所属区域，
- 生成的大图要能正好打印在A4纸上


以下是一个使用 Go 语言编写的程序，满足您的所有要求。

### 程序说明

1.  **依赖库**：使用了 Go 官方扩展图像处理库 `golang.org/x/image/draw` 来实现高质量的图片缩放。
2.  **A4 尺寸计算**：程序按照 300 DPI（打印标准）计算 A4 纸的像素尺寸（2480 x 3508 像素）。
3.  **布局逻辑**：
    *   将 A4 纸划分为 4行 3列 的网格。
    *   预留了页边距和图片间的间隙。
    *   **缩放策略**：计算每个格子的可用空间，将图片等比例缩放以适应该空间（保持宽高比，不会拉伸变形），然后居中放置。
    *   **边框**：在图片绘制前，先绘制一个比图片稍大的淡灰色矩形作为边框。

### 前置准备

在运行代码前，需要安装图像处理库：

```bash
go get golang.org/x/image/draw
```

### Go 代码

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw" // 高质量缩放库
)

// 配置参数
const (
	Rows       = 4                         // 行数
	Cols       = 3                         // 列数
	ImagesPerSheet = Rows * Cols           // 每张图包含的小图数量

	// A4纸尺寸 (像素, 按300DPI标准)
	// 210mm * 300 / 25.4 ≈ 2480
	// 297mm * 300 / 25.4 ≈ 3508
	A4WidthPx  = 2480
	A4HeightPx = 3508

	Margin     = 60  // 页边距 (像素)
	Gap        = 30  // 图片之间的间隙 (像素)
	BorderSize = 4   // 边框宽度 (像素)
)

// borderColor 淡色边框颜色 (浅灰色)
var borderColor = color.RGBA{200, 200, 200, 255}
var bgColor = color.RGBA{255, 255, 255, 255} // 白色背景

func main() {
	// 1. 搜索当前目录下的图片文件
	extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	var imageFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过子目录和生成的输出文件
		if info.IsDir() || strings.HasPrefix(strings.ToLower(info.Name()), "output_") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("当前目录下未找到图片文件。")
		return
	}

	fmt.Printf("共找到 %d 张图片，开始处理...\n", len(imageFiles))

	// 2. 分批处理图片 (每12张一组)
	for i := 0; i < len(imageFiles); i += ImagesPerSheet {
		end := i + ImagesPerSheet
		if end > len(imageFiles) {
			end = len(imageFiles)
		}
		batch := imageFiles[i:end]

		outputFileName := fmt.Sprintf("output_%d.jpg", i/ImagesPerSheet+1)
		err := createContactSheet(batch, outputFileName)
		if err != nil {
			log.Printf("生成 %s 失败: %v", outputFileName, err)
		} else {
			fmt.Printf("已生成: %s (包含 %d 张图片)\n", outputFileName, len(batch))
		}
	}

	fmt.Println("处理完成。")
}

// createContactSheet 生成一张拼接的大图
func createContactSheet(files []string, outputFilename string) error {
	// 创建 A4 尺寸的画布
	canvas := image.NewRGBA(image.Rect(0, 0, A4WidthPx, A4HeightPx))
	
	// 填充白色背景
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 计算每个单元格的尺寸
	// 总宽度 - 左右边距 - (列数-1)*间隙 = 可用宽度
	cellWidth := (A4WidthPx - 2*Margin - (Cols-1)*Gap) / Cols
	cellHeight := (A4HeightPx - 2*Margin - (Rows-1)*Gap) / Rows

	for idx, file := range files {
		// 打开图片文件
		imgFile, err := os.Open(file)
		if err != nil {
			log.Printf("无法打开文件 %s: %v", file, err)
			continue
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			log.Printf("无法解码文件 %s: %v", file, err)
			continue
		}

		// 计算当前图片在网格中的位置 (行, 列)
		row := idx / Cols
		col := idx % Cols

		// 计算单元格左上角的坐标
		cellX := Margin + col*(cellWidth+Gap)
		cellY := Margin + row*(cellHeight+Gap)
		cellRect := image.Rect(cellX, cellY, cellX+cellWidth, cellY+cellHeight)

		// 将图片绘制到单元格中 (保持比例，缩放并居中)
		drawImageFit(canvas, img, cellRect)
	}

	// 保存结果
	outFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: 95})
}

// drawImageFit 将图片缩放并居中绘制到指定区域，保持宽高比，并添加边框
func drawImageFit(dst draw.Image, src image.Image, cellRect image.Rectangle) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	cellW := cellRect.Dx()
	cellH := cellRect.Dy()

	// 1. 计算缩放比例，保持宽高比
	scale := float64(cellW) / float64(srcW)
	if int(float64(srcH)*scale) > cellH {
		scale = float64(cellH) / float64(srcH)
	}

	// 计算缩放后的尺寸
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	// 2. 计算居中位置
	dstX := cellRect.Min.X + (cellW-newW)/2
	dstY := cellRect.Min.Y + (cellH-newH)/2

	// 图片绘制的目标区域
	imgRect := image.Rect(dstX, dstY, dstX+newW, dstY+newH)

	// 3. 绘制淡色边框 (在图片区域稍微向外扩展一点，作为边框)
	borderRect := image.Rect(
		imgRect.Min.X-BorderSize,
		imgRect.Min.Y-BorderSize,
		imgRect.Max.X+BorderSize,
		imgRect.Max.Y+BorderSize,
	)
	draw.Draw(dst, borderRect, &image.Uniform{borderColor}, image.Point{}, draw.Src)

	// 4. 缩放并绘制图片 (使用 CatmullRom 插值算法，质量较高)
	// 这里缩放源图像到新尺寸
	scaledImg := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), src, srcBounds, draw.Over, nil)

	// 将缩放后的图片画到画布上
	draw.Draw(dst, imgRect, scaledImg, image.Point{}, draw.Over)
}

// 辅助函数：用于获取文件的读取流，忽略 io 参数（本例中直接用 os.Open）
var _ = io.EOF
```

### 如何运行

1.  将上述代码保存为 `main.go`。
2.  在当前目录下放置一些测试图片（jpg/png）。
3.  在终端运行：
    ```bash
    go run main.go
    ```
4.  程序运行后会在当前目录生成 `output_1.jpg`, `output_2.jpg` 等文件。

### 关键实现细节

*   **A4 尺寸适配**：使用了打印行业标准 300 DPI，将物理尺寸（210mm x 297mm）转换为像素，确保打印清晰。
*   **高质量缩放**：使用了 `golang.org/x/image/draw` 包中的 `CatmullRom` 插值算法。这比标准库 `image/draw` 的最近邻缩放效果要好得多，避免了缩放后的锯齿感。
*   **保持比例**：代码逻辑先计算目标区域（单元格）的宽高比和原图的宽高比，通过比较决定以“宽度适配”还是“高度适配”为准，确保图片填满区域且不变形。
*   **边框与间隙**：先绘制一个比缩放后图片稍大的淡灰色矩形作为边框，再绘制图片，实现了视觉上的分隔效果。


# .\main.go:15:5: draw redeclared in this block
.\main.go:7:5: other declaration of draw
.\main.go:15:5: "golang.org/x/image/draw" imported and not used
.\main.go:178:10: undefined: draw.CatmullRom

这是一个包名冲突的问题。Go 语言中标准库 `image/draw` 和扩展库 `golang.org/x/image/draw` 的包名相同，导致编译器报错。

解决方法是为扩展库起一个别名（例如 `xdraw`）。

以下是修复后的完整代码：

### 修复后的代码 (main.go)

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw" // 标准库 draw
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw" // 扩展库 draw，使用别名 xdraw
)

// 配置参数
const (
	Rows           = 4     // 行数
	Cols           = 3     // 列数
	ImagesPerSheet = Rows * Cols // 每张图包含的小图数量

	// A4纸尺寸 (像素, 按300DPI标准)
	// 210mm * 300 / 25.4 ≈ 2480
	// 297mm * 300 / 25.4 ≈ 3508
	A4WidthPx  = 2480
	A4HeightPx = 3508

	Margin     = 60  // 页边距 (像素)
	Gap        = 30  // 图片之间的间隙 (像素)
	BorderSize = 4   // 边框宽度 (像素)
)

// borderColor 淡色边框颜色 (浅灰色)
var borderColor = color.RGBA{200, 200, 200, 255}
var bgColor = color.RGBA{255, 255, 255, 255} // 白色背景

func main() {
	// 1. 搜索当前目录下的图片文件
	extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	var imageFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过子目录和生成的输出文件
		if info.IsDir() || strings.HasPrefix(strings.ToLower(info.Name()), "output_") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("当前目录下未找到图片文件。")
		return
	}

	fmt.Printf("共找到 %d 张图片，开始处理...\n", len(imageFiles))

	// 2. 分批处理图片 (每12张一组)
	for i := 0; i < len(imageFiles); i += ImagesPerSheet {
		end := i + ImagesPerSheet
		if end > len(imageFiles) {
			end = len(imageFiles)
		}
		batch := imageFiles[i:end]

		outputFileName := fmt.Sprintf("output_%d.jpg", i/ImagesPerSheet+1)
		err := createContactSheet(batch, outputFileName)
		if err != nil {
			log.Printf("生成 %s 失败: %v", outputFileName, err)
		} else {
			fmt.Printf("已生成: %s (包含 %d 张图片)\n", outputFileName, len(batch))
		}
	}

	fmt.Println("处理完成。")
}

// createContactSheet 生成一张拼接的大图
func createContactSheet(files []string, outputFilename string) error {
	// 创建 A4 尺寸的画布
	canvas := image.NewRGBA(image.Rect(0, 0, A4WidthPx, A4HeightPx))

	// 填充白色背景 (使用标准库 draw)
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 计算每个单元格的尺寸
	// 总宽度 - 左右边距 - (列数-1)*间隙 = 可用宽度
	cellWidth := (A4WidthPx - 2*Margin - (Cols-1)*Gap) / Cols
	cellHeight := (A4HeightPx - 2*Margin - (Rows-1)*Gap) / Rows

	for idx, file := range files {
		// 打开图片文件
		imgFile, err := os.Open(file)
		if err != nil {
			log.Printf("无法打开文件 %s: %v", file, err)
			continue
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			log.Printf("无法解码文件 %s: %v", file, err)
			continue
		}

		// 计算当前图片在网格中的位置 (行, 列)
		row := idx / Cols
		col := idx % Cols

		// 计算单元格左上角的坐标
		cellX := Margin + col*(cellWidth+Gap)
		cellY := Margin + row*(cellHeight+Gap)
		cellRect := image.Rect(cellX, cellY, cellX+cellWidth, cellY+cellHeight)

		// 将图片绘制到单元格中 (保持比例，缩放并居中)
		drawImageFit(canvas, img, cellRect)
	}

	// 保存结果
	outFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: 95})
}

// drawImageFit 将图片缩放并居中绘制到指定区域，保持宽高比，并添加边框
func drawImageFit(dst draw.Image, src image.Image, cellRect image.Rectangle) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	cellW := cellRect.Dx()
	cellH := cellRect.Dy()

	// 1. 计算缩放比例，保持宽高比
	scale := float64(cellW) / float64(srcW)
	if int(float64(srcH)*scale) > cellH {
		scale = float64(cellH) / float64(srcH)
	}

	// 计算缩放后的尺寸
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	// 2. 计算居中位置
	dstX := cellRect.Min.X + (cellW-newW)/2
	dstY := cellRect.Min.Y + (cellH-newH)/2

	// 图片绘制的目标区域
	imgRect := image.Rect(dstX, dstY, dstX+newW, dstY+newH)

	// 3. 绘制淡色边框 (在图片区域稍微向外扩展一点，作为边框)
	borderRect := image.Rect(
		imgRect.Min.X-BorderSize,
		imgRect.Min.Y-BorderSize,
		imgRect.Max.X+BorderSize,
		imgRect.Max.Y+BorderSize,
	)
	// 使用标准库 draw 绘制纯色矩形
	draw.Draw(dst, borderRect, &image.Uniform{borderColor}, image.Point{}, draw.Src)

	// 4. 缩放并绘制图片 (使用扩展库 xdraw 进行高质量缩放)
	// CatmullRom 插值算法质量较高
	xdraw.CatmullRom.Scale(dst, imgRect, src, srcBounds, draw.Over, nil)
}

// 辅助函数：用于获取文件的读取流，忽略 io 参数（本例中直接用 os.Open）
var _ = io.EOF
```

### 修改说明

1.  **引入别名**：
    将 `golang.org/x/image/draw` 修改为 `xdraw "golang.org/x/image/draw"`。
    这样标准库 `image/draw` 保持原名 `draw`，用于基础绘图（如画背景、画边框矩形）。
    扩展库使用别名 `xdraw`，专门用于高质量缩放 `xdraw.CatmullRom.Scale`。

2.  **优化缩放逻辑**：
    在 `drawImageFit` 函数中，直接使用 `xdraw.CatmullRom.Scale` 将原图缩放并绘制到目标画布上，不再需要手动创建中间的 `scaledImg` 变量，代码更简洁高效。

您可以直接复制上述代码覆盖原文件，然后再次运行 `go run main.go` 即可。


# 淡色边框应该加在区域上，而不是加在图片上

您是对的。之前的逻辑是根据图片缩放后的尺寸画边框，导致边框大小不一且紧贴图片。

修改后的逻辑是：**边框是固定的，直接画在网格区域（`cellRect`）的边缘**。图片则缩放后居中绘制在这个边框区域内部。这样无论图片比例如何，每个区域的边框都是整齐划一的。

以下是修正后的代码：

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// 配置参数
const (
	Rows           = 4
	Cols           = 3
	ImagesPerSheet = Rows * Cols

	// A4纸尺寸 (像素, 按300DPI标准)
	A4WidthPx  = 2480
	A4HeightPx = 3508

	Margin     = 60  // 页边距
	Gap        = 30  // 图片之间的间隙
	BorderSize = 4   // 边框宽度
)

// borderColor 淡色边框颜色 (浅灰色)
var borderColor = color.RGBA{220, 220, 220, 255} // 稍微调淡一点
var bgColor = color.RGBA{255, 255, 255, 255}     // 白色背景

func main() {
	extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	var imageFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(strings.ToLower(info.Name()), "output_") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("当前目录下未找到图片文件。")
		return
	}

	fmt.Printf("共找到 %d 张图片，开始处理...\n", len(imageFiles))

	for i := 0; i < len(imageFiles); i += ImagesPerSheet {
		end := i + ImagesPerSheet
		if end > len(imageFiles) {
			end = len(imageFiles)
		}
		batch := imageFiles[i:end]

		outputFileName := fmt.Sprintf("output_%d.jpg", i/ImagesPerSheet+1)
		err := createContactSheet(batch, outputFileName)
		if err != nil {
			log.Printf("生成 %s 失败: %v", outputFileName, err)
		} else {
			fmt.Printf("已生成: %s (包含 %d 张图片)\n", outputFileName, len(batch))
		}
	}
	fmt.Println("处理完成。")
}

func createContactSheet(files []string, outputFilename string) error {
	canvas := image.NewRGBA(image.Rect(0, 0, A4WidthPx, A4HeightPx))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	cellWidth := (A4WidthPx - 2*Margin - (Cols-1)*Gap) / Cols
	cellHeight := (A4HeightPx - 2*Margin - (Rows-1)*Gap) / Rows

	for idx, file := range files {
		imgFile, err := os.Open(file)
		if err != nil {
			log.Printf("无法打开文件 %s: %v", file, err)
			continue
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			log.Printf("无法解码文件 %s: %v", file, err)
			continue
		}

		row := idx / Cols
		col := idx % Cols

		cellX := Margin + col*(cellWidth+Gap)
		cellY := Margin + row*(cellHeight+Gap)
		
		// 单元格区域
		cellRect := image.Rect(cellX, cellY, cellX+cellWidth, cellY+cellHeight)

		drawImageFit(canvas, img, cellRect)
	}

	outFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: 95})
}

// drawImageFit 逻辑修改：
// 1. 先绘制区域边框（淡色矩形填充整个区域）
// 2. 再将图片缩放居中绘制在该区域之上
func drawImageFit(dst draw.Image, src image.Image, cellRect image.Rectangle) {
	// 1. 在区域上绘制淡色边框/背景
	// 这里直接用淡色填充整个 cellRect，作为底色和边框
	draw.Draw(dst, cellRect, &image.Uniform{borderColor}, image.Point{}, draw.Src)

	// 2. 计算图片缩放比例，保持宽高比
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	cellW := cellRect.Dx()
	cellH := cellRect.Dy()

	// 计算缩放比例
	scale := float64(cellW) / float64(srcW)
	if int(float64(srcH)*scale) > cellH {
		scale = float64(cellH) / float64(srcH)
	}

	// 计算缩放后的尺寸
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	// 3. 计算居中位置
	// 图片绘制在区域内部，居中对齐
	dstX := cellRect.Min.X + (cellW-newW)/2
	dstY := cellRect.Min.Y + (cellH-newH)/2

	// 图片绘制的目标区域
	imgRect := image.Rect(dstX, dstY, dstX+newW, dstY+newH)

	// 4. 缩放并绘制图片 (覆盖在淡色边框之上)
	// 如果图片比例和区域不一致，四周会露出步骤1绘制的淡色边框
	xdraw.CatmullRom.Scale(dst, imgRect, src, srcBounds, draw.Over, nil)
}

var _ = io.EOF
```

### 主要修改点

1.  **边框绘制位置**：
    *   **原逻辑**：根据图片缩放后的尺寸计算边框，导致边框大小随图片变化。
    *   **新逻辑**：直接在 `cellRect`（网格分配的固定区域）上绘制淡色矩形。这个矩形既是背景也是边框。

2.  **视觉层级**：
    *   先绘制淡色矩形作为底层。
    *   再将图片缩放并居中绘制在矩形之上。
    *   这样，当图片比例与区域不一致时，露出的部分（上下或左右空白）会呈现为淡色边框效果，整齐划一。

### 效果
现在，无论图片是长条形还是正方形，它们所在的每一个网格区域都会有一个统一的淡色边框，整体排版看起来更加整齐、正式，适合打印。


# 第一个问题，我需要的是加边框线，而不是用背景色表示边框。
第二个问题，当生成多张大图时，第二张以后的大图都是空白

感谢您的反馈。这两个问题我都已经定位并修复：

1.  **边框问题**：之前的代码是填充整个背景作为"底色"，现在修改为绘制矩形线条（DrawRect 逻辑），只画轮廓，中间是透明的。
2.  **空白图片问题**：这是一个经典的 Go 语言图片处理陷阱。代码中导入了 `image/jpeg` 用于生成图片，但**忘记导入 `image/png` 和 `image/gif`**。Go 语言默认只支持程序中显式导入的格式。如果第二批图片包含 PNG 文件，解码会失败，导致空白。现在已添加所有常用格式的隐式导入。

以下是修正后的完整代码：

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// 必须导入这些包以注册 PNG/GIF 解码器，否则无法读取这些格式
	_ "image/gif"
	_ "image/png"

	xdraw "golang.org/x/image/draw"
)

// 配置参数
const (
	Rows           = 4
	Cols           = 3
	ImagesPerSheet = Rows * Cols

	// A4纸尺寸 (像素, 按300DPI标准)
	A4WidthPx  = 2480
	A4HeightPx = 3508

	Margin     = 60  // 页边距
	Gap        = 30  // 图片之间的间隙
	BorderSize = 2   // 边框线条宽度 (像素)
)

// borderColor 淡色边框颜色 (浅灰色)
var borderColor = color.RGBA{180, 180, 180, 255}
var bgColor = color.RGBA{255, 255, 255, 255} // 白色背景

func main() {
	// 1. 搜索当前目录下的图片文件
	extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	var imageFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过子目录和生成的输出文件
		if info.IsDir() || strings.HasPrefix(strings.ToLower(info.Name()), "output_") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("当前目录下未找到图片文件。")
		return
	}

	// 排序文件，确保顺序固定
	sort.Strings(imageFiles)

	fmt.Printf("共找到 %d 张图片，开始处理...\n", len(imageFiles))

	// 2. 分批处理图片 (每12张一组)
	for i := 0; i < len(imageFiles); i += ImagesPerSheet {
		end := i + ImagesPerSheet
		if end > len(imageFiles) {
			end = len(imageFiles)
		}
		batch := imageFiles[i:end]

		outputFileName := fmt.Sprintf("output_%d.jpg", i/ImagesPerSheet+1)
		err := createContactSheet(batch, outputFileName)
		if err != nil {
			log.Printf("生成 %s 失败: %v", outputFileName, err)
		} else {
			fmt.Printf("已生成: %s (包含 %d 张图片)\n", outputFileName, len(batch))
		}
	}

	fmt.Println("处理完成。")
}

// createContactSheet 生成一张拼接的大图
func createContactSheet(files []string, outputFilename string) error {
	// 创建 A4 尺寸的画布
	canvas := image.NewRGBA(image.Rect(0, 0, A4WidthPx, A4HeightPx))

	// 填充白色背景
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 计算每个单元格的尺寸
	cellWidth := (A4WidthPx - 2*Margin - (Cols-1)*Gap) / Cols
	cellHeight := (A4HeightPx - 2*Margin - (Rows-1)*Gap) / Rows

	for idx, file := range files {
		// 打开图片文件
		imgFile, err := os.Open(file)
		if err != nil {
			log.Printf("无法打开文件 %s: %v", file, err)
			continue
		}

		// 解码图片 (依赖 import _ "image/png" 等来支持不同格式)
		img, format, err := image.Decode(imgFile)
		imgFile.Close() // 及时关闭文件句柄
		
		if err != nil {
			log.Printf("无法解码文件 %s (格式: %s): %v", file, format, err)
			continue
		}

		// 计算当前图片在网格中的位置 (行, 列)
		row := idx / Cols
		col := idx % Cols

		// 计算单元格左上角的坐标
		cellX := Margin + col*(cellWidth+Gap)
		cellY := Margin + row*(cellHeight+Gap)
		cellRect := image.Rect(cellX, cellY, cellX+cellWidth, cellY+cellHeight)

		// 将图片绘制到单元格中
		drawImageFit(canvas, img, cellRect)
	}

	// 保存结果
	outFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, canvas, &jpeg.Options{Quality: 95})
}

// drawImageFit 绘制边框线，并将图片缩放居中绘制
func drawImageFit(dst draw.Image, src image.Image, cellRect image.Rectangle) {
	// 1. 绘制淡色边框线
	// Go 的 draw 库没有直接的 DrawRect，我们通过画四个矩形来实现线条
	// 上边框
	draw.Draw(dst, image.Rect(cellRect.Min.X, cellRect.Min.Y, cellRect.Max.X, cellRect.Min.Y+BorderSize), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// 下边框
	draw.Draw(dst, image.Rect(cellRect.Min.X, cellRect.Max.Y-BorderSize, cellRect.Max.X, cellRect.Max.Y), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// 左边框
	draw.Draw(dst, image.Rect(cellRect.Min.X, cellRect.Min.Y, cellRect.Min.X+BorderSize, cellRect.Max.Y), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// 右边框
	draw.Draw(dst, image.Rect(cellRect.Max.X-BorderSize, cellRect.Min.Y, cellRect.Max.X, cellRect.Max.Y), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	// 2. 计算图片在区域内的可用空间 (减去边框线占用的空间)
	padding := BorderSize + 2 // 额外加 2 像素留白，防止图片紧贴边框
	innerRect := image.Rect(
		cellRect.Min.X+padding,
		cellRect.Min.Y+padding,
		cellRect.Max.X-padding,
		cellRect.Max.Y-padding,
	)
	innerW := innerRect.Dx()
	innerH := innerRect.Dy()

	// 3. 计算缩放比例，保持宽高比
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	if srcW <= 0 || srcH <= 0 {
		return // 防止除零错误
	}

	scale := float64(innerW) / float64(srcW)
	if int(float64(srcH)*scale) > innerH {
		scale = float64(innerH) / float64(srcH)
	}

	// 计算缩放后的尺寸
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	// 4. 计算居中位置
	dstX := innerRect.Min.X + (innerW-newW)/2
	dstY := innerRect.Min.Y + (innerH-newH)/2

	// 图片绘制的目标区域
	imgRect := image.Rect(dstX, dstY, dstX+newW, dstY+newH)

	// 5. 缩放并绘制图片
	xdraw.CatmullRom.Scale(dst, imgRect, src, srcBounds, draw.Over, nil)
}

var _ = io.EOF
```

### 主要修改点详解

1.  **修复“空白图片”问题**：
    *   在 `import` 区域添加了 `_ "image/png"` 和 `_ "image/gif"`。
    *   Go 语言的 `image.Decode` 函数需要预先注册解码器才能识别特定格式。如果不导入这些包，程序只能读取 JPEG，遇到 PNG 就会报错并跳过，导致该位置空白。
    *   加入了 `sort.Strings(imageFiles)` 确保每次运行顺序一致，方便调试。

2.  **修复“边框线”问题**：
    *   删除了之前的“填充整个区域作为背景”的逻辑。
    *   在 `drawImageFit` 开头增加了绘制四条边的逻辑：通过绘制四个细长的矩形（上下左右）来组成一个空心的边框线条。
    *   增加了 `padding` 变量，在计算图片位置时，向内收缩了边框宽度+2像素，确保图片不会覆盖边框线，且与边框保持微小距离。

您可以直接运行这段代码，现在生成的图片应该有清晰的灰色边框线，并且能正确处理 JPG/PNG/GIF 格式的图片了。
