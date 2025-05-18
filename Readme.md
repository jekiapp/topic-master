# Overview 

This architecture pursues the high modularity of codes and also testability and maintainability to achieve high-frequency delivery. The code example is written in Go. However, the same principle would be applicable to another language.

The ideal situation for this architecture to take benefit is for a development team consisting of **3–7 people**, with a deployment schedule of **1–2 times** per week.

The components of this architecture are split into two categories: **Infrastructure** & **Business Logic**. The high-level dependency direction can be seen in the diagram below:

![dependency-diagram.png](assets%2Fdependency-diagram.png)

# Usage

For a detailed usage and guideline, please [go to this article](https://manusia-serba-tanya.medium.com/high-modular-architecture-specification-guideline-41a29779ba91).